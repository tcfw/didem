package cli

import (
	"bytes"
	"context"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	apipb "github.com/tcfw/didem/api"
	"github.com/tcfw/didem/internal/api"
	"github.com/tcfw/didem/internal/utils/logging"
)

var (
	emCmd = &cobra.Command{
		Use:   "em",
		Short: "Email commands",
	}

	em_sendCmd = &cobra.Command{
		Use:   "send",
		Short: "Send an email",
		Run:   runEmSend,
	}
)

func init() {
	em_sendCmd.Flags().String("from", "", "identity to send as. blank defaults to the default identity")
	em_sendCmd.Flags().StringArrayP("file", "f", []string{}, "parts to read in as the email content. Can be used multiple times. Use '-' for stdin")
	em_sendCmd.Flags().StringArrayP("header", "H", []string{}, "headers to append to the email in the format KEY=value. Can be used multiple times")
	em_sendCmd.Flags().StringArrayP("to", "t", []string{}, "identity to send to")
	em_sendCmd.Flags().StringP("subject", "s", "", "email subject")
	em_sendCmd.Flags().String("stdin-split", "", "split stdin with specified delimiter")
}

func runEmSend(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	api, err := api.NewClient()
	if err != nil {
		logging.WithError(err).Error("constructing client")
		return
	}

	subject, _ := cmd.Flags().GetString("subject")
	from, _ := cmd.Flags().GetString("from")
	to, err := cmd.Flags().GetStringArray("to")
	if err != nil {
		logging.WithError(err).Error("failed to understand flag 'to'")
		return
	}
	headers, err := cmd.Flags().GetStringArray("header")
	if err != nil {
		logging.WithError(err).Error("failed to understand flag 'header'")
		return
	}
	parts, err := cmd.Flags().GetStringArray("file")
	if err != nil {
		logging.WithError(err).Error("failed to understand flag 'file'")
		return
	}
	stdinDelim, _ := cmd.Flags().GetString("stdin-split")

	req := &apipb.EmSendRequest{
		Headers: map[string]string{},
		To:      to,
		From:    from,
	}

	if subject != "" {
		req.Headers["subject"] = subject
	}

	for _, h := range headers {
		hParts := strings.SplitN(h, "=", 1)
		if len(hParts) != 2 {
			logging.Entry().Error("header should have 2 logical parts")
			return
		}
		req.Headers[hParts[0]] = hParts[1]
	}

	var readStdIn bool
	for _, p := range parts {
		var datas [][]byte
		if p == "-" {
			if readStdIn {
				logging.Entry().Error("stdin specified mulitple times and already read in")
				return
			}
			datas, err = readPartsFromStdin([]byte(stdinDelim))
			readStdIn = true
			if err != nil {
				logging.WithError(err).Error("failed to read part")
			}
		} else {
			data, err := readPartFromFile(p)
			if err != nil {
				logging.WithError(err).Error("failed to read part")
			}
			datas = [][]byte{data}
		}

		for _, data := range datas {
			mime := detectMIME(data)

			part := &apipb.EmSendPart{
				Mime: mime,
				Data: data,
			}

			req.Parts = append(req.Parts, part)
		}
	}

	if viper.GetBool("verbose") {
		logging.Entry().WithFields(logrus.Fields{
			"to":      req.To,
			"headers": req.Headers,
			"from":    req.From,
		}).Info("Sending email")
		for _, p := range req.Parts {
			ent := logging.Entry().WithField("mime", p.Mime)
			if strings.HasPrefix(p.Mime, "application/octect-stream") {
				ent = ent.WithField("b64-data", string(firstNBytes([]byte(base64.StdEncoding.EncodeToString(p.Data)), 20)))
			} else {
				ent = ent.WithField("data", string(firstNBytes(p.Data, 20)))
			}
			ent.Info("part")
		}
	}

	_, err = api.Em().Send(ctx, req)
	if err != nil {
		logging.WithError(err).Error("failed to send")
	}

	return
}

func firstNBytes(d []byte, n int) []byte {
	if len(d) < n {
		n = len(d)
	}

	return d[:n]
}

func detectMIME(b []byte) string {
	return http.DetectContentType(b)
}

func readPartFromFile(f string) ([]byte, error) {
	return ioutil.ReadFile(f)
}

func readPartsFromStdin(delim []byte) ([][]byte, error) {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, errors.Wrap(err, "reading stdin")
	}

	if len(delim) == 0 {
		return [][]byte{data}, nil
	}

	return bytes.Split(data, delim), nil
}

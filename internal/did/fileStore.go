package did

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/did"
	"gopkg.in/yaml.v3"
)

type IdentityFileStore struct {
	Ids []IdentityFileStoreId `yaml:"ids"`
}

type IdentityFileStoreId struct {
	Type string `yaml:"type"`
	Data string `yaml:"data"`
}

var _ did.IdentityStore = (*FileStore)(nil)

type FileStore struct {
	path string
	ids  IdentityFileStore
	idx  map[string]did.PrivateIdentity

	mu sync.Mutex
}

func NewFileStore(path string) (*FileStore, error) {
	f := &FileStore{path: path}
	if err := f.read(); err != nil {
		return nil, err
	}

	return f, nil
}

func (fs *FileStore) read() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	f, err := os.OpenFile(fs.path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return errors.Wrap(err, "opening identity file for read")
	}
	defer f.Close()

	d, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "")
	}

	if err := yaml.Unmarshal(d, &fs.ids); err != nil {
		return errors.Wrap(err, "unmarshalling identity data")
	}

	return fs.buildIdx()
}

func (fs *FileStore) buildIdx() error {
	//assumes locked fs.mu

	fs.idx = make(map[string]did.PrivateIdentity, len(fs.ids.Ids))

	for _, fid := range fs.ids.Ids {
		id, err := fs.decodeType(fid.Type, fid.Data)
		if err != nil {
			return errors.Wrap(err, "decoding id")
		}

		pid, err := id.PublicIdentity()
		if err != nil {
			return errors.Wrap(err, "getting public id from private")
		}

		fs.idx[pid.ID] = id
	}

	return nil
}

func (fs *FileStore) decodeType(t string, data string) (did.PrivateIdentity, error) {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, errors.Wrap(err, "decoding b64 identity data")
	}

	switch t {
	case "ed25519":
		return did.NewEd25519Identity(raw), nil
	default:
		return nil, fmt.Errorf("unknown key type %s", t)
	}
}

func (fs *FileStore) Add(id did.PrivateIdentity) error {
	pid, err := id.PublicIdentity()
	if err != nil {
		return err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	//check if in idx
	if _, ok := fs.idx[pid.ID]; ok {
		return nil
	}

	switch t := id.(type) {
	case *did.Ed25519Identity:
		id := id.(*did.Ed25519Identity)
		pk := id.PrivateKey().(ed25519.PrivateKey)
		raw := base64.StdEncoding.EncodeToString([]byte(pk))
		f := IdentityFileStoreId{Type: "ed25519", Data: raw}
		fs.ids.Ids = append(fs.ids.Ids, f)
	default:
		return fmt.Errorf("unknown did PK type %T", t)
	}

	return fs.write()
}

func (fs *FileStore) write() error {

	f, err := os.OpenFile(fs.path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return errors.Wrap(err, "opening identity file for write")
	}
	defer f.Close()

	d, err := yaml.Marshal(&fs.ids)
	if err != nil {
		return errors.Wrap(err, "marshalling identity data")
	}

	f.Truncate(0)
	_, err = f.Write(d)
	return err
}

func (fs *FileStore) Find(id string) (did.PrivateIdentity, error) {
	i, ok := fs.idx[id]
	if !ok {
		return nil, errors.New("not found")
	}

	return i, nil
}

func (fs *FileStore) List() ([]did.PrivateIdentity, error) {
	ids := make([]did.PrivateIdentity, 0, len(fs.idx))

	for _, id := range fs.idx {
		ids = append(ids, id)
	}

	return ids, nil
}

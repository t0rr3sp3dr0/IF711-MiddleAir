package crypto

import (
	"bytes"
	"compress/gzip"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"github.com/t0rr3sp3dr0/middleair/util"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	_ "golang.org/x/crypto/ripemd160"
)

type SecureConn struct {
	net.Conn
	publicKey       *packet.PublicKey
	privateKey      *packet.PrivateKey
	publicEntities  openpgp.EntityList
	privateEntities openpgp.EntityList
	sharedKey       []byte
}

func NewSecureConn(conn net.Conn) (*SecureConn, error) {
	e := &SecureConn{
		Conn: conn,
	}

	w := &util.WrapperConn{
		Conn: conn,
	}

	pub, priv, err := func() (*packet.PublicKey, *packet.PrivateKey, error) {
		pubFile, err := os.Open(os.Getenv("MIDDLEAIR_PUBKEY"))
		if err != nil {
			return nil, nil, err
		}
		defer pubFile.Close()

		privFile, err := os.Open(os.Getenv("MIDDLEAIR_PRIVKEY"))
		if err != nil {
			return nil, nil, err
		}
		defer privFile.Close()

		pubBlock, err := armor.Decode(pubFile)
		if err != nil {
			return nil, nil, err
		}

		privBlock, err := armor.Decode(privFile)
		if err != nil {
			return nil, nil, err
		}

		if pubBlock.Type != openpgp.PublicKeyType {
			return nil, nil, errors.New("Invalid Public Key File")
		}

		if privBlock.Type != openpgp.PrivateKeyType {
			return nil, nil, errors.New("Invalid Private Key File")
		}

		pubPacket, err := packet.NewReader(pubBlock.Body).Next()
		if err != nil {
			return nil, nil, err
		}

		privPacket, err := packet.NewReader(privBlock.Body).Next()
		if err != nil {
			return nil, nil, err
		}

		pubKey, ok := pubPacket.(*packet.PublicKey)
		if !ok {
			return nil, nil, errors.New("Invalid Public Key")
		}

		privKey, ok := privPacket.(*packet.PrivateKey)
		if !ok {
			return nil, nil, errors.New("Invalid Private Key")
		}

		return pubKey, privKey, nil
	}()

	key, err := func() (*rsa.PrivateKey, error) {
		if err == nil {
			return nil, nil
		}
		return rsa.GenerateKey(rand.Reader, 4096)
	}()
	if err != nil {
		return nil, err
	}

	publicKeyBuf := bytes.NewBuffer(nil)
	publicKey := func() *packet.PublicKey {
		if pub != nil {
			return pub
		}
		return packet.NewRSAPublicKey(time.Now(), &key.PublicKey)
	}()
	if err := publicKey.Serialize(publicKeyBuf); err != nil {
		return nil, err
	}

	if _, err := w.WriteData(publicKeyBuf.Bytes()); err != nil {
		return nil, err
	}

	pk, err := w.ReadData()
	if err != nil {
		return nil, err
	}
	pkt, err := packet.NewReader(bytes.NewBuffer(pk)).Next()
	if err != nil {
		return nil, err
	}
	e.publicKey = pkt.(*packet.PublicKey)

	e.privateKey = func() *packet.PrivateKey {
		if pub != nil {
			return priv
		}
		return packet.NewRSAPrivateKey(time.Now(), key)
	}()

	e.publicEntities = []*openpgp.Entity{
		newEntity(e.publicKey, nil),
	}

	e.privateEntities = []*openpgp.Entity{
		newEntity(publicKey, e.privateKey),
	}

	e.sharedKey = make([]byte, 512)
	if _, err := rand.Read(e.sharedKey); err != nil {
		return nil, err
	}

	sharedKeyBuf := bytes.NewBuffer(nil)
	postCompressed, err := gzip.NewWriterLevel(sharedKeyBuf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	oncePostCompressed := &sync.Once{}
	defer oncePostCompressed.Do(func() {
		if err := postCompressed.Close(); err != nil {
		}
	})
	plaintext, err := openpgp.Encrypt(postCompressed, e.publicEntities, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	oncePlaintext := &sync.Once{}
	defer oncePostCompressed.Do(func() {
		if err := plaintext.Close(); err != nil {
		}
	})
	preCompressed, err := gzip.NewWriterLevel(plaintext, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	oncePreCompressed := &sync.Once{}
	defer oncePostCompressed.Do(func() {
		if err := preCompressed.Close(); err != nil {
		}
	})
	if _, err := io.Copy(preCompressed, bytes.NewReader(e.sharedKey)); err != nil {
		return nil, err
	}
	if err := preCompressed.Close(); err != nil {
		return nil, err
	}
	oncePreCompressed.Do(func() {})
	if err := plaintext.Close(); err != nil {
		return nil, err
	}
	oncePlaintext.Do(func() {})
	if err := postCompressed.Close(); err != nil {
		return nil, err
	}
	oncePostCompressed.Do(func() {})

	if _, err := w.WriteData(sharedKeyBuf.Bytes()); err != nil {
		return nil, err
	}

	sk, err := w.ReadData()
	if err != nil {
		return nil, err
	}

	preDecompressed, err := gzip.NewReader(bytes.NewReader(sk))
	if err != nil {
		return nil, err
	}
	defer preDecompressed.Close()
	md, err := openpgp.ReadMessage(preDecompressed, e.privateEntities, nil, nil)
	if err != nil {
		return nil, err
	}
	postDecompressed, err := gzip.NewReader(md.UnverifiedBody)
	if err != nil {
		return nil, err
	}
	defer postDecompressed.Close()
	skBuf := bytes.NewBuffer(nil)
	if _, err := io.Copy(skBuf, postDecompressed); err != nil {
		return nil, err
	}

	for i, b := range skBuf.Bytes() {
		e.sharedKey[i] = e.sharedKey[i] ^ b
	}

	return e, nil
}

func newEntity(pubKey *packet.PublicKey, privKey *packet.PrivateKey) *openpgp.Entity {
	userId := packet.NewUserId("", "", "")
	config := packet.Config{
		RSABits:                4096,
		DefaultHash:            crypto.SHA256,
		DefaultCipher:          packet.CipherAES256,
		DefaultCompressionAlgo: packet.CompressionZLIB,
		CompressionConfig: &packet.CompressionConfig{
			Level: 9,
		},
		Time: func() time.Time {
			return time.Unix(0, 0)
		},
	}
	e := openpgp.Entity{
		PrimaryKey: pubKey,
		PrivateKey: privKey,
		Identities: map[string]*openpgp.Identity{
			userId.Id: &openpgp.Identity{
				Name:   userId.Name,
				UserId: userId,
				SelfSignature: &packet.Signature{
					SigType:      packet.SigTypePositiveCert,
					PubKeyAlgo:   packet.PubKeyAlgoRSA,
					Hash:         config.Hash(),
					CreationTime: config.Now(),
					IssuerKeyId:  &pubKey.KeyId,
					IsPrimaryId:  new(bool),
					FlagsValid:   true,
					FlagCertify:  true,
					FlagSign:     true,
				},
			},
		},
		Subkeys: []openpgp.Subkey{
			openpgp.Subkey{
				PublicKey:  pubKey,
				PrivateKey: privKey,
				Sig: &packet.Signature{
					SigType:                   packet.SigTypeSubkeyBinding,
					PubKeyAlgo:                packet.PubKeyAlgoRSA,
					Hash:                      config.Hash(),
					CreationTime:              config.Now(),
					IssuerKeyId:               &pubKey.KeyId,
					FlagsValid:                true,
					FlagEncryptCommunications: true,
					FlagEncryptStorage:        true,
				},
			},
		},
	}
	return &e
}

func (e *SecureConn) ReadData() ([]byte, error) {
	buf := make([]byte, math.MaxInt16)
	n, err := e.Read(buf)
	if n < 8 || err != nil {
		return nil, err
	}
	size := int(binary.LittleEndian.Uint64(buf[:8]))

	preDecompressed, err := gzip.NewReader(bytes.NewBuffer(buf[8 : size+8]))
	if err != nil {
		return nil, err
	}
	defer preDecompressed.Close()
	md, err := openpgp.ReadMessage(preDecompressed, nil, func([]openpgp.Key, bool) ([]byte, error) { return e.sharedKey, nil }, nil)
	if err != nil {
		return nil, err
	}
	postDecompressed, err := gzip.NewReader(md.UnverifiedBody)
	if err != nil {
		return nil, err
	}
	defer postDecompressed.Close()
	decBuf := bytes.NewBuffer(nil)
	if _, err := io.Copy(decBuf, postDecompressed); err != nil {
		return nil, err
	}

	return decBuf.Bytes(), nil
}

func (e *SecureConn) WriteData(data []byte) (int, error) {
	encBuf := bytes.NewBuffer(nil)
	postCompressed, err := gzip.NewWriterLevel(encBuf, gzip.BestSpeed)
	if err != nil {
		return -1, err
	}
	oncePostCompressed := &sync.Once{}
	defer oncePostCompressed.Do(func() {
		if err := postCompressed.Close(); err != nil {
		}
	})
	plaintext, err := openpgp.SymmetricallyEncrypt(postCompressed, e.sharedKey, nil, nil)
	if err != nil {
		return -1, err
	}
	oncePlaintext := &sync.Once{}
	defer oncePlaintext.Do(func() {
		if err := plaintext.Close(); err != nil {
		}
	})
	preCompressed, err := gzip.NewWriterLevel(plaintext, gzip.BestSpeed)
	if err != nil {
		return -1, err
	}
	oncePreCompressed := &sync.Once{}
	defer oncePostCompressed.Do(func() {
		if err := preCompressed.Close(); err != nil {
		}
	})
	if _, err := io.Copy(preCompressed, bytes.NewBuffer(data)); err != nil {
		return -1, err
	}
	if err := preCompressed.Close(); err != nil {
		return -1, err
	}
	oncePreCompressed.Do(func() {})
	if err := plaintext.Close(); err != nil {
		return -1, err
	}
	oncePlaintext.Do(func() {})
	if err := postCompressed.Close(); err != nil {
		return -1, err
	}
	oncePostCompressed.Do(func() {})

	data = encBuf.Bytes()

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(len(data)))
	bytes = append(bytes, data...)

	return e.Write(bytes)
}

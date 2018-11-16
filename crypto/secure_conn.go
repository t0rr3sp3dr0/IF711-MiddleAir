package crypto

import (
	"bytes"
	"compress/gzip"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"../util"
	"golang.org/x/crypto/openpgp"
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

func NewSecureConn(conn net.Conn) (_ *SecureConn, ret error) {
	e := &SecureConn{
		Conn: conn,
	}

	w := &util.WrapperConn{
		Conn: conn,
	}

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
		return nil, err
	}

	publicKeyBuf := bytes.NewBuffer(nil)
	publicKey := packet.NewRSAPublicKey(time.Now(), &key.PublicKey)
	if err := publicKey.Serialize(publicKeyBuf); err != nil {
		panic(err)
		return nil, err
	}

	if _, err := w.WriteData(publicKeyBuf.Bytes()); err != nil {
		panic(err)
		return nil, err
	}

	pk, err := w.ReadData()
	if err != nil {
		panic(err)
		return nil, err
	}
	pkt, err := packet.NewReader(bytes.NewBuffer(pk)).Next()
	if err != nil {
		panic(err)
		return nil, err
	}
	e.publicKey = pkt.(*packet.PublicKey)

	e.privateKey = packet.NewRSAPrivateKey(time.Now(), key)

	e.publicEntities = []*openpgp.Entity{
		newEntity(e.publicKey, nil),
	}

	e.privateEntities = []*openpgp.Entity{
		newEntity(publicKey, e.privateKey),
	}

	e.sharedKey = make([]byte, 512)
	if _, err := rand.Read(e.sharedKey); err != nil {
		panic(err)
		return nil, err
	}

	sharedKeyBuf := bytes.NewBuffer(nil)
	postCompressed, err := gzip.NewWriterLevel(sharedKeyBuf, gzip.BestCompression)
	if err != nil {
		panic(err)
		return nil, err
	}
	oncePostCompressed := &sync.Once{}
	defer oncePostCompressed.Do(func() {
		if err := postCompressed.Close(); err != nil {
			panic(err)
		}
	})
	plaintext, err := openpgp.Encrypt(postCompressed, e.publicEntities, nil, nil, nil)
	if err != nil {
		panic(err)
		return nil, err
	}
	oncePlaintext := &sync.Once{}
	defer oncePostCompressed.Do(func() {
		if err := plaintext.Close(); err != nil {
			panic(err)
		}
	})
	preCompressed, err := gzip.NewWriterLevel(plaintext, gzip.BestCompression)
	if err != nil {
		panic(err)
		return nil, err
	}
	oncePreCompressed := &sync.Once{}
	defer oncePostCompressed.Do(func() {
		if err := preCompressed.Close(); err != nil {
			panic(err)
		}
	})
	if _, err := io.Copy(preCompressed, bytes.NewReader(e.sharedKey)); err != nil {
		panic(err)
		return nil, err
	}
	if err := preCompressed.Close(); err != nil {
		panic(err)
		return nil, err
	}
	oncePreCompressed.Do(func() {})
	if err := plaintext.Close(); err != nil {
		panic(err)
		return nil, err
	}
	oncePlaintext.Do(func() {})
	if err := postCompressed.Close(); err != nil {
		panic(err)
		return nil, err
	}
	oncePostCompressed.Do(func() {})

	if _, err := w.WriteData(sharedKeyBuf.Bytes()); err != nil {
		panic(err)
		return nil, err
	}

	sk, err := w.ReadData()
	if err != nil {
		panic(err)
		return nil, err
	}

	preDecompressed, err := gzip.NewReader(bytes.NewReader(sk))
	if err != nil {
		panic(err)
		return nil, err
	}
	defer preDecompressed.Close()
	md, err := openpgp.ReadMessage(preDecompressed, e.privateEntities, nil, nil)
	if err != nil {
		panic(err)
		return nil, err
	}
	postDecompressed, err := gzip.NewReader(md.UnverifiedBody)
	if err != nil {
		panic(err)
		return nil, err
	}
	defer postDecompressed.Close()
	skBuf := bytes.NewBuffer(nil)
	if _, err := io.Copy(skBuf, postDecompressed); err != nil {
		panic(err)
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
					PreferredHash:             []uint8{8}, // SHA-256
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
			panic(err)
		}
	})
	plaintext, err := openpgp.SymmetricallyEncrypt(postCompressed, e.sharedKey, nil, nil)
	if err != nil {
		return -1, err
	}
	oncePlaintext := &sync.Once{}
	defer oncePlaintext.Do(func() {
		if err := plaintext.Close(); err != nil {
			panic(err)
		}
	})
	preCompressed, err := gzip.NewWriterLevel(plaintext, gzip.BestSpeed)
	if err != nil {
		return -1, err
	}
	oncePreCompressed := &sync.Once{}
	defer oncePostCompressed.Do(func() {
		if err := preCompressed.Close(); err != nil {
			panic(err)
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

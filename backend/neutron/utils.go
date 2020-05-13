package neutron

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/yaoice/cni-ipam-neutron/backend/allocator"
	"log"
	"strings"
)

var (
	keyText = "aabcice12798akljzmknm.ahkjkljl;k"
	commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
)

func connectStore(openstackConf *allocator.OpenStackConf) (*gophercloud.ServiceClient, error) {
	// keystone auth v3
	if strings.HasSuffix(openstackConf.AuthUrl, "v3") {
		return connectWithKeyStoneV3(openstackConf)
	} else {
		return connectWithKeyStoneV2(openstackConf)
	}
}


// Keystone v2
func connectWithKeyStoneV2(openstackConf *allocator.OpenStackConf) (*gophercloud.ServiceClient, error) {
	// keystone auth v2
	provider, err := openstack.AuthenticatedClient(gophercloud.AuthOptions{
		IdentityEndpoint: openstackConf.AuthUrl,
		Username: openstackConf.UserName,
		Password: encrypt.aesDecrypt(openstackConf.PassWord),
		TenantName: openstackConf.Project,
	})
	if err != nil {
		return nil, err
	}
	networkClient, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		log.Printf("Get network client err: %v", err)
		return nil, err
	}
	return networkClient, err
}


// Keystone v3
func connectWithKeyStoneV3(openstackConf *allocator.OpenStackConf) (*gophercloud.ServiceClient, error) {
	provider, err := openstack.AuthenticatedClient(gophercloud.AuthOptions{
		IdentityEndpoint: openstackConf.AuthUrl,
		Username: openstackConf.UserName,
		Password: encrypt.aesDecrypt(openstackConf.PassWord),
		TenantName: openstackConf.Project,
		DomainName: openstackConf.Domain,
	})
	if err != nil {
		return nil, err
	}
	networkClient, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		log.Printf("Get network client err: %v", err)
		return nil, err
	}
	return networkClient, err
}

var encrypt = NewEncrypter()

func NewEncrypter() *encrypter {
	// create encrypt algorithm
	cip, err := aes.NewCipher([]byte(keyText))
	if err != nil {
		log.Printf("Get aes cipher err: %v", err)
		return nil
	}
	return &encrypter{
		cip: cip,
		commonIV: commonIV,
	}
}

type encrypter struct {
	cip cipher.Block
	commonIV []byte
}

// AES Encrypt
func (e *encrypter) aesEncrypt(plainText string) string {
	plainTextByte := []byte(plainText)
	// encrypt plaintext
	cfb := cipher.NewCFBEncrypter(e.cip, e.commonIV)
	cipherText := make([]byte, len(plainTextByte))
	cfb.XORKeyStream(cipherText, plainTextByte)
	return hex.EncodeToString(cipherText)
}

// AES Decrypt
func (e *encrypter) aesDecrypt(cipherText string) string {
	cipherTextByte, err := hex.DecodeString(cipherText)
	if err != nil {
		panic(err)
	}
	// decrypt cipherText
	cfbdec := cipher.NewCFBDecrypter(e.cip, e.commonIV)
	plaintextCopy := make([]byte, len(cipherTextByte))
	cfbdec.XORKeyStream(plaintextCopy, cipherTextByte)
	return string(plaintextCopy)
}

// return bool pointer
func getBoolPointer(b bool) *bool {
	return &b
}

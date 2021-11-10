package utility

import (
	"crypto/sha256"
	"encoding/hex"
	"talisman/gitrepo"
)

type SHA256Hasher interface {
	CollectiveSHA256Hash(paths []string) string
	Start() error
	Shutdown() error
}

type DefaultSHA256Hasher struct{}

//CollectiveSHA256Hash return collective sha256 hash of the passed paths
func (*DefaultSHA256Hasher) CollectiveSHA256Hash(paths []string) string {
	return collectiveSHA256Hash(paths, SafeReadFile)
}

func (*DefaultSHA256Hasher) Start() error    { return nil }
func (*DefaultSHA256Hasher) Shutdown() error { return nil }

type gitBatchSHA256Hasher struct {
	br gitrepo.BatchReader
}

func (g *gitBatchSHA256Hasher) CollectiveSHA256Hash(paths []string) string {
	return collectiveSHA256Hash(paths, g.br.Read)
}

func (g *gitBatchSHA256Hasher) Start() error {
	return g.br.Start()
}

func (g *gitBatchSHA256Hasher) Shutdown() error {
	return g.br.Shutdown()
}

func hashByte(contentPtr *[]byte) string {
	contents := *contentPtr
	hasher := sha256.New()
	hasher.Write(contents)
	return hex.EncodeToString(hasher.Sum(nil))
}

func collectiveSHA256Hash(paths []string, FileReader func(string) ([]byte, error)) string {
	var finHash = ""
	for _, path := range paths {
		sbyte := []byte(finHash)
		concatBytes := hashByte(&sbyte)
		nameByte := []byte(path)
		nameHash := hashByte(&nameByte)
		fileBytes, _ := FileReader(path)
		fileHash := hashByte(&fileBytes)
		finHash = concatBytes + fileHash + nameHash
	}
	c := []byte(finHash)
	m := hashByte(&c)
	return m
}

//MakeHasher returns a SHA256 file/object hasher based on mode and a repo root
func MakeHasher(mode string, root string) SHA256Hasher {
	var hasher SHA256Hasher
	switch mode {
	case "pre-push":
		hasher = &gitBatchSHA256Hasher{gitrepo.NewBatchGitHeadPathReader(root)}
	case "pre-commit":
		hasher = &gitBatchSHA256Hasher{gitrepo.NewBatchGitStagedPathReader(root)}
	case "scan":
		hasher = &gitBatchSHA256Hasher{gitrepo.NewBatchGitObjectHashReader(root)}
	case "pattern":
		hasher = &DefaultSHA256Hasher{}
	case "checksum":
		hasher = &gitBatchSHA256Hasher{gitrepo.NewBatchGitHeadPathReader(root)}
	case "default":
		hasher = &DefaultSHA256Hasher{}
	}
	return hasher
}

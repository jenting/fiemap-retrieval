package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/longhorn/sparse-tools/sparse"
	"github.com/sirupsen/logrus"
)

const (
	MB = int64(1024 * 1024)
	GB = MB * 1024

	srcPrefix    = "ssync-src"
	testFileName = srcPrefix + "-fiemap-10gb-file"
	testFileSize = 10 * GB

	testHoleBlockSize = int64(4096)
	testDataBlockSize = int64(4096)
	// testDataBlockSize = int64(128*MB - testHoleBlockSize)
)

func main() {
	srcPath := filepath.Join(os.TempDir(), testFileName)

	logrus.Infof("Write sparse file to %s with size %v (block %v, hole %v)", srcPath, testFileSize, testDataBlockSize, testHoleBlockSize)

	if err := writeMultipleHolesData(srcPath, testFileSize, testDataBlockSize, testHoleBlockSize); err != nil {
		logrus.Fatalf("failed to create fiemap test file path: %v error: %v", srcPath, err)
	}
}

func writeMultipleHolesData(filePath string, fileSize int64, dataSize int64, holeSize int64) (err error) {
	if fileSize%(dataSize+holeSize) != 0 {
		return fmt.Errorf("fileSize %v needs to be a multiple of dataSize %v + holeSize %v", fileSize, dataSize, holeSize)
	}

	const GB = int64(1024 * 1024 * 1024)
	sizeInGB := fileSize / GB
	logrus.Infof("Start to create a %vGB file with multiple hole", sizeInGB)
	f, err := sparse.NewDirectFileIoProcessor(filePath, os.O_RDWR, 0666, true)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = os.Remove(filePath)
		}
	}()
	defer f.Close()
	defer f.Sync()
	if err := f.Truncate(fileSize); err != nil {
		return err
	}

	startTime := time.Now()
	deltaTime := time.Now()

	// random is pretty slow, if called in the loop below for each character
	// better to call it once per block, so we get a full block of a single random character
	for offset := int64(0); offset < fileSize; {
		blockData := randomBlock(dataSize)
		if nw, err := f.WriteAt(blockData, offset); err != nil {
			return fmt.Errorf("Write at %v, number of write %v, error: %v", offset, nw, err)
		}
		offset += dataSize
		if err := NewFiemapFile(f.GetFile()).PunchHole(offset, holeSize); err != nil {
			return fmt.Errorf("Punch hole at %v error: %v", offset, err)
		}
		offset += holeSize

		if offset%GB == 0 {
			writtenGB := offset / GB
			logrus.Infof("Wrote %vGB of %vGB time delta: %.2f time elapsed: %.2f",
				writtenGB, sizeInGB,
				time.Now().Sub(deltaTime).Seconds(),
				time.Now().Sub(startTime).Seconds())
			deltaTime = time.Now()
		}
	}

	logrus.Infof("Done creating a %vGB file with multiple hole, time elapsed: %.2f", sizeInGB, time.Now().Sub(startTime).Seconds())
	return nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomBlock(n int64) []byte {
	char := letterBytes[rand.Intn(len(letterBytes))]
	b := make([]byte, n)
	for i := range b {
		b[i] = char
	}
	return b
}

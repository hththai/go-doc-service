package item

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
)

type Item struct {
	GUID   string
	Id     string
	Name   string
	Path   string
	Format string
}

func NewItem(name string) *Item {
	return &Item{
		GUID:   uuid.New().String(),
		Name:   name,
		Path:   "",
		Id:     "001",
		Format: "docx",
	}
}

func SayHello() {
	fmt.Printf("hello")
}

func ShowItem(item Item) {
	fmt.Printf("Item is::: %s\n", item.GUID)
	fmt.Printf("Item Name is::: %s\n ", item.Name)
}

func ShowItemAsJson(item Item) {
	jsonData, err := json.MarshalIndent(item, "", " ")
	if err != nil {
		fmt.Printf("Error converting to JSON::: ", err)
		return
	}

	fmt.Printf(string(jsonData))
}

func SaveToFileID(item Item) {
	id := item.Id
	err := os.WriteFile("filedata/filedata.id", []byte(id), 0644)

	if err != nil {
		fmt.Println("Error Writing file::: ", err)
		return
	}

	fmt.Println("ID saved to filedata.id")
}

func SaveCollectionFileToFileID(items []Item) {
	jsonData, err := json.MarshalIndent(items, "", " ")
	if err != nil {
		fmt.Println("Error marshaling Json:::", err)
		return
	}

	err = os.WriteFile("filedata/filedata.id", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing file::: ", err)
		return
	}
	fmt.Println("Document records are saved")
}

func ReadFileID() {
	data, err := os.ReadFile("filedata/filedata.id")

	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	var readDocs []Item
	err = json.Unmarshal(data, &readDocs)

	if err != nil {
		fmt.Println("Error unmarshaling JSON::", err)
		return
	}

	fmt.Println("Documents read from file:::")
	for _, doc := range readDocs {
		fmt.Printf("ID: %s | Format: %s\n", doc.GUID, doc.Format)
	}

}

func encryptFile(inputPath, outputPath string, key []byte) error {
	// Read file content
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Use GCM mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Generate nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// Encrypt
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	// Write encrypted data
	return os.WriteFile(outputPath, ciphertext, 0644)
}

func decryptFile(inputPath string, key []byte) ([]byte, error) {
	ciphertext, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

func main() {
	// item := NewItem("Example")
	// fmt.Printf("Item::: %+v\n", item)
	items := []Item{
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
	}

	jsonData, err := json.MarshalIndent(items, "", " ")
	if err != nil {
		fmt.Println("Error marshaling Json:::", err)
		return
	}

	err = os.WriteFile("filedata/filedata.id", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing file::: ", err)
		return
	}
	fmt.Println("Document records are saved")
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Структура для обработки ответа от Vault
type ResponseContentSupplier struct {
	Data map[string]interface{} `json:"data"`
}

var (
	vaultUrl   = "https://URL/v1/kv/data"
	vaultToken = "TOKEN"
	client     = &http.Client{}
)

// Функция для запроса данных из Vault
func get(name string) (map[string]interface{}, error) {
	vaultUrl := fmt.Sprintf("%s/%s", vaultUrl, name)
	req, err := http.NewRequest("GET", vaultUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", vaultToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("invalid response status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var httpResponse ResponseContentSupplier

	if err := json.Unmarshal(body, &httpResponse); err != nil {
		return nil, err
	}

	return httpResponse.Data, nil
}

// Запись данных в файл YAML с нужными преобразованиями
func writeToFileAsYaml(data map[string]interface{}, filename string) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	yamlString := string(yamlData)

	// Ищем позицию строки "metadata=" и обрезаем всё, что после неё
	if idx := strings.Index(yamlString, "metadata:"); idx != -1 {
		yamlString = yamlString[:idx]
	}
	// Удаляем строку "data:" и всё, что идёт после "metadata:"
	yamlString = strings.ReplaceAll(yamlString, "data:\n", "")

	// Заменяем двоеточие на знак равенства и убираем начальные пробелы
	yamlString = strings.ReplaceAll(yamlString, ": ", "=")
	yamlString = regexp.MustCompile(`(?m)^\s+`).ReplaceAllString(yamlString, "")

	yamlString = strings.TrimSpace(yamlString) // Удаляем пустые строки в начале и конце

	// Записываем преобразованный YAML в файл
	if err := os.WriteFile(filename, []byte(yamlString), 0644); err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: v2file <secretPath> <fileName>")
		os.Exit(1)
	}

	secretPath := os.Args[1]
	fileName := os.Args[2]

	data, err := get(secretPath)
	if err != nil {
		log.Fatalf("Error getting data from Vault: %v", err)
	}

	fmt.Println("Data received from Vault: \n", data, "\n")

	if err := writeToFileAsYaml(data, fileName); err != nil {
		log.Fatalf("Error writing data to file: %v", err)
	}

	fmt.Printf("Data successfully written to %s\n", fileName)
}

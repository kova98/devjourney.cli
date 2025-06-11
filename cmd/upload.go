package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/adrg/frontmatter"
	"github.com/kova98/devjourney/cli/cfg"
	"github.com/kova98/devjourney/cli/models"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var apiKey string

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload markdown entry and all linked resources",
	Long:  `TODO`,
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs, fileExists),
	Run:   upload,
}

func fileExists(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no file provided")
	}
	file := args[0]
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", file)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Your API key (required)")
	err := rootCmd.MarkPersistentFlagRequired("api-key")
	if err != nil {
		fmt.Println("Error marking api-key as required:", err)
		return
	}
}

func check(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func upload(cmd *cobra.Command, args []string) {
	filePath := args[0]
	if filePath == "" {
		fmt.Println("Please provide a file to upload.")
		return
	}

	// Read the whole file content
	c, err := os.ReadFile(filePath)
	markdown := string(c)
	check(err)

	// Get all media
	mdImageRegex := regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`)
	htmlImgRegex := regexp.MustCompile(`<img[^>]+src="([^">]+)"`)
	htmlVideoRegex := regexp.MustCompile(`<video[^>]+src="([^">]+)"|<source[^>]+src="([^">]+)"`)
	mdVideoRegex := regexp.MustCompile(`\[[^\]]*\]\(([^)]+\.(mp4|webm|ogg))\)`)
	matches := [][]string{}
	matches = append(matches, mdImageRegex.FindAllStringSubmatch(markdown, -1)...)
	matches = append(matches, htmlImgRegex.FindAllStringSubmatch(markdown, -1)...)
	matches = append(matches, htmlVideoRegex.FindAllStringSubmatch(markdown, -1)...)
	matches = append(matches, mdVideoRegex.FindAllStringSubmatch(markdown, -1)...)

	fmt.Printf("Processing file: %s\n", filePath)

	// Call DevJourney POST /content for each media file

	for _, match := range matches {
		if len(match) < 1 {
			fmt.Println("Invalid image tag: " + match[0])
			continue
		}

		path := match[1]
		absPath := filepath.Join(filepath.Dir(filePath), path)

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			fmt.Println("File not found:", absPath)
			continue
		}

		url := fmt.Sprintf("%s/content", cfg.ApiRoot)
		req := formFileRequest(absPath, url)
		client := &http.Client{}
		resp, err := client.Do(req)
		check(err)

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Bad status: %d\n", resp.StatusCode)
			os.Exit(1)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		check(err)

		var result struct {
			Uri string
		}
		err = json.Unmarshal(bodyBytes, &result)
		check(err)

		// Replace urls in markdown with uploaded content URLs
		oldTag := match[0]
		newTag := strings.Replace(oldTag, path, result.Uri, 1)
		markdown = strings.Replace(markdown, oldTag, newTag, -1)

		err = resp.Body.Close()
		check(err)
	}

	// Extract metadata from markdown
	type Metadata struct {
		Date      string `yaml:"date"`
		Project   string `yaml:"project"`
		Title     string `yaml:"title"`
		Mood      string `yaml:"mood"`
		TimeSpent string `yaml:"time-spent"`
	}

	var meta Metadata
	_, err = frontmatter.Parse(strings.NewReader(markdown), &meta)
	if err != nil {
		log.Fatal(err)
	}

	// Call DevJourney POST /entries

	if meta.TimeSpent == "" {
		fmt.Println("Time spent is required in the metadata.")
		os.Exit(1)
	}

	timeSpent, err := time.ParseDuration(meta.TimeSpent)
	check(err)
	date, err := time.Parse(time.DateOnly, meta.Date)
	check(err)

	userInfo := getUserInfo()
	projectId := ""
	for _, project := range userInfo.Projects {
		if project.Slug == meta.Project {
			projectId = project.ID
			break
		}
	}

	// TODO: replace projectId with project slug
	req := &models.CreateEntryRequest{
		ProjectID:    projectId,
		Date:         date,
		MinutesSpent: int(math.Round(timeSpent.Minutes())),
		Mood:         meta.Mood,
		Content:      "",
		Title:        meta.Title,
	}

	url := fmt.Sprintf("%s/entries", cfg.ApiRoot)

	body, err := json.Marshal(req)
	check(err)

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	check(err)

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	check(err)

	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Bad status: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	defer resp.Body.Close()

	var result struct {
		ID string `json:"id"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	check(err)

	fmt.Printf("Entry created successfully with ID: %s\n", result.ID)
}

func getUserInfo() models.GetUserInfoResponse {
	url := fmt.Sprintf("%s/users/info", cfg.ApiRoot)
	req, err := http.NewRequest("GET", url, nil)
	check(err)

	req.Header.Set("x-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	check(err)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Bad status: %d\n", resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Println("Response body:", string(bodyBytes))
		os.Exit(1)
	}

	defer resp.Body.Close()

	var userInfo models.GetUserInfoResponse
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	check(err)

	return userInfo
}

func formFileRequest(path string, url string) *http.Request {
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", file.Name())
	check(err)

	_, err = io.Copy(part, file)
	check(err)

	err = writer.Close()
	check(err)

	req, err := http.NewRequest("POST", url, body)
	check(err)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("x-api-key", apiKey)

	return req
}

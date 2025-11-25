package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	APIKey      = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJHcm91cE5hbWUiOiLpuL_lkIjnp5HmioDogqHku73mnInpmZDlhazlj7giLCJVc2VyTmFtZSI6Ium4v-WQiOenkeaKgOiCoeS7veaciemZkOWFrOWPuCIsIkFjY291bnQiOiIiLCJTdWJqZWN0SUQiOiIxOTI1MTE4NTM3NjUxNzI1MTE5IiwiUGhvbmUiOiIxMzQ3NDM3MTE5MyIsIkdyb3VwSUQiOiIxOTI1MTE4NTM3NjQzMzM2NTExIiwiUGFnZU5hbWUiOiIiLCJNYWlsIjoiIiwiQ3JlYXRlVGltZSI6IjIwMjUtMTEtMjUgMTc6MjI6MjUiLCJUb2tlblR5cGUiOjEsImlzcyI6Im1pbmltYXgifQ.JlAlyt7bImSvJNIRj1PjTJdJ6JwbAjS2mRlDwdmzsv00pml4g2lFR0jhxwYwmxuNbgustjWTB--KXnDJFpg1WzQZbQkMsxGGSZjmiqWxOqwi9MpLJsncBa7teYM4yzHBpE2BsFVIT2baJM_nADSBO68oojjnNWXmTEhlW4Z75B-tG9JC07n_G-o3gF422T_ot6IOSdngLPbzZ6kgu_hrs7Sori24UetxsZimFHcWFiZG4ctMDn1Ia_DxgVH39HsIlbBQpZ3eQwqUZyyT0YyQ-XUOiMsd4y6ubewsN3kNrR943Ro7WwlVaBbQCxK0np0dapMjgYRr6K-Qni3Z-3-Pmw"
	BaseURL     = "https://api.minimaxi.com/v1/t2a_async_v2"
	QueryURL    = "https://api.minimaxi.com/v1/query/t2a_async_query_v2?task_id=%s"
	RetrieveURL = "https://api.minimaxi.com/v1/files/retrieve?file_id=%s"
	Model       = "speech-02-hd"
	VoiceID     = "Chinese (Mandarin)_Warm_Bestie"
)

type VoiceSetting struct {
	VoiceID string  `json:"voice_id"`
	Speed   float64 `json:"speed"`
	Vol     float64 `json:"vol"`
	Pitch   int     `json:"pitch"`
}

type AudioSetting struct {
	AudioSampleRate int    `json:"audio_sample_rate"`
	Bitrate         int    `json:"bitrate"`
	Format          string `json:"format"`
	Channel         int    `json:"channel"`
}

type T2ARequest struct {
	Model        string       `json:"model"`
	Text         string       `json:"text"`
	VoiceSetting VoiceSetting `json:"voice_setting"`
	AudioSetting AudioSetting `json:"audio_setting"`
}

type ResponseWrapper struct {
	BaseResp struct {
		StatusCode int    `json:"status_code"`
		StatusMsg  string `json:"status_msg"`
	} `json:"base_resp"`
}

type T2AResponse struct {
	ResponseWrapper
	TaskID int64 `json:"task_id"`
}

type QueryResponse struct {
	ResponseWrapper
	Status string `json:"status"` // "Processing", "Success", "Failed"
	FileID int64  `json:"file_id"`
}

type RetrieveResponse struct {
	ResponseWrapper
	File struct {
		DownloadURL string `json:"download_url"`
	} `json:"file"`
}

func main() {
	// Read text from 1.md
	content, err := os.ReadFile("../data/1.md")
	if err != nil {
		fmt.Printf("Error reading 1.md: %v\n", err)
		return
	}
	text := string(content)
	if text == "" {
		fmt.Println("Warning: 1.md is empty. Using default text.")
		text = "Hello! This is a test of the MiniMax speech synthesis API. We are generating this audio asynchronously."
	}

	// Allow command line override
	if len(os.Args) > 1 {
		text = os.Args[1]
	}

	fmt.Printf("Generating audio for text (length %d): %s...\n", len(text), text[:min(len(text), 50)])

	taskID, err := generateAudio(text)
	if err != nil {
		fmt.Printf("Error generating audio: %v\n", err)
		return
	}
	fmt.Printf("Task initiated. Task ID: %d\n", taskID)

	// Poll for status
	var fileID int64
	for {
		status, fid, err := queryStatus(taskID)
		if err != nil {
			fmt.Printf("Error querying status: %v\n", err)
			return
		}
		fmt.Printf("Status: %s\n", status)

		if status == "Success" {
			fileID = fid
			break
		} else if status == "Failed" {
			fmt.Println("Audio generation failed.")
			return
		}

		time.Sleep(2 * time.Second)
	}

	fmt.Printf("Generation successful. File ID: %d\n", fileID)

	// Get download URL
	downloadURL, err := getDownloadURL(fileID)
	if err != nil {
		fmt.Printf("Error getting download URL: %v\n", err)
		return
	}
	fmt.Printf("Download URL: %s\n", downloadURL)

	// Download file
	tarFilename := "../data/output.tar"
	err = downloadFile(downloadURL, tarFilename)
	if err != nil {
		fmt.Printf("Error downloading file: %v\n", err)
		return
	}
	fmt.Printf("Downloaded archive to %s. Extracting...\n", tarFilename)

	// Extract tar
	mp3File, err := extractTar(tarFilename, "../data")
	if err != nil {
		fmt.Printf("Error extracting tar: %v\n", err)
		return
	}

	if mp3File != "" {
		fmt.Printf("Audio extracted to: %s\n", mp3File)
		// Rename to output.mp3
		err := os.Rename(mp3File, "../data/output.mp3")
		if err != nil {
			fmt.Printf("Error renaming file: %v\n", err)
		} else {
			fmt.Println("Renamed to output.mp3")
		}
	} else {
		fmt.Println("No MP3 file found in the archive.")
	}
}

func generateAudio(text string) (int64, error) {
	reqBody := T2ARequest{
		Model: Model,
		Text:  text,
		VoiceSetting: VoiceSetting{
			VoiceID: VoiceID,
			Speed:   0.8,
			Vol:     1.0,
			Pitch:   0,
		},
		AudioSetting: AudioSetting{
			AudioSampleRate: 32000,
			Bitrate:         128000,
			Format:          "mp3",
			Channel:         1,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	var t2aResp T2AResponse
	if err := json.Unmarshal(bodyBytes, &t2aResp); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %v, body: %s", err, string(bodyBytes))
	}

	if t2aResp.BaseResp.StatusCode != 0 {
		return 0, fmt.Errorf("API Error: %s (Code: %d)", t2aResp.BaseResp.StatusMsg, t2aResp.BaseResp.StatusCode)
	}

	return t2aResp.TaskID, nil
}

func queryStatus(taskID int64) (string, int64, error) {
	url := fmt.Sprintf(QueryURL, fmt.Sprintf("%d", taskID))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Authorization", "Bearer "+APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	var queryResp QueryResponse
	if err := json.Unmarshal(bodyBytes, &queryResp); err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal response: %v, body: %s", err, string(bodyBytes))
	}

	if queryResp.BaseResp.StatusCode != 0 {
		return "", 0, fmt.Errorf("API Error: %s (Code: %d)", queryResp.BaseResp.StatusMsg, queryResp.BaseResp.StatusCode)
	}

	return queryResp.Status, queryResp.FileID, nil
}

func getDownloadURL(fileID int64) (string, error) {
	url := fmt.Sprintf(RetrieveURL, fmt.Sprintf("%d", fileID))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	var retrieveResp RetrieveResponse
	if err := json.Unmarshal(bodyBytes, &retrieveResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v, body: %s", err, string(bodyBytes))
	}

	if retrieveResp.BaseResp.StatusCode != 0 {
		return "", fmt.Errorf("API Error: %s (Code: %d)", retrieveResp.BaseResp.StatusMsg, retrieveResp.BaseResp.StatusCode)
	}

	return retrieveResp.File.DownloadURL, nil
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractTar(tarPath, destDir string) (string, error) {
	file, err := os.Open(tarPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	tr := tar.NewReader(file)
	var mp3Path string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		target := filepath.Join(destDir, header.Name)

		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", err
			}
		} else if header.Typeflag == tar.TypeReg {
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", err
			}

			f, err := os.Create(target)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return "", err
			}
			f.Close()

			if strings.HasSuffix(header.Name, ".mp3") {
				mp3Path = target
			}
		}
	}
	return mp3Path, nil
}

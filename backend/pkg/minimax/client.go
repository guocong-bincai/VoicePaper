package minimax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 这里复用你之前的 API Key 和逻辑，但封装得更好
const (
	APIKey      = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJHcm91cE5hbWUiOiLpuL_lkIjnp5HmioDogqHku73mnInpmZDlhazlj7giLCJVc2VyTmFtZSI6Ium4v-WQiOenkeaKgOiCoeS7veaciemZkOWFrOWPuCIsIkFjY291bnQiOiIiLCJTdWJqZWN0SUQiOiIxOTI1MTE4NTM3NjUxNzI1MTE5IiwiUGhvbmUiOiIxMzQ3NDM3MTE5MyIsIkdyb3VwSUQiOiIxOTI1MTE4NTM3NjQzMzM2NTExIiwiUGFnZU5hbWUiOiIiLCJNYWlsIjoiIiwiQ3JlYXRlVGltZSI6IjIwMjUtMTEtMjUgMTc6MjI6MjUiLCJUb2tlblR5cGUiOjEsImlzcyI6Im1pbmltYXgifQ.JlAlyt7bImSvJNIRj1PjTJdJ6JwbAjS2mRlDwdmzsv00pml4g2lFR0jhxwYwmxuNbgustjWTB--KXnDJFpg1WzQZbQkMsxGGSZjmiqWxOqwi9MpLJsncBa7teYM4yzHBpE2BsFVIT2baJM_nADSBO68oojjnNWXmTEhlW4Z75B-tG9JC07n_G-o3gF422T_ot6IOSdngLPbzZ6kgu_hrs7Sori24UetxsZimFHcWFiZG4ctMDn1Ia_DxgVH39HsIlbBQpZ3eQwqUZyyT0YyQ-XUOiMsd4y6ubewsN3kNrR943Ro7WwlVaBbQCxK0np0dapMjgYRr6K-Qni3Z-3-Pmw"
	BaseURL     = "https://api.minimaxi.com/v1/t2a_async_v2"
	QueryURL    = "https://api.minimaxi.com/v1/query/t2a_async_query_v2?task_id=%s"
	RetrieveURL = "https://api.minimaxi.com/v1/files/retrieve?file_id=%s"
	Model       = "speech-02-hd"
	VoiceID     = "Chinese (Mandarin)_Warm_Bestie"
)

// GenerateSpeech 是对外的统一接口，处理所有异步轮询逻辑，直接返回音频二进制数据
func GenerateSpeech(text string) ([]byte, error) {
	// 1. 发起请求
	taskID, err := initiateTask(text)
	if err != nil {
		return nil, err
	}

	// 2. 轮询状态
	var fileID int64
	for {
		status, fid, err := queryStatus(taskID)
		if err != nil {
			return nil, err
		}
		if status == "Success" {
			fileID = fid
			break
		} else if status == "Failed" {
			return nil, fmt.Errorf("generation failed on server side")
		}
		time.Sleep(2 * time.Second)
	}

	// 3. 获取下载链接
	downloadURL, err := getDownloadURL(fileID)
	if err != nil {
		return nil, err
	}

	// 4. 下载并返回数据 (注意：这里简化了处理，实际可能需要解压 tar)
	// MiniMax 返回的是 tar 包吗？根据之前的代码，是的。
	// 所以我们需要下载 -> 解压 -> 提取 mp3
	return downloadAndExtract(downloadURL)
}

// ... (辅助函数 initiateTask, queryStatus, getDownloadURL, downloadAndExtract)
// 为了节省篇幅，我先写核心逻辑，具体实现可以复用之前的 main.go 代码
// 这里需要把之前的 main.go 里的 struct 和 helper functions 搬过来

type T2ARequest struct {
	Model        string       `json:"model"`
	Text         string       `json:"text"`
	VoiceSetting VoiceSetting `json:"voice_setting"`
	AudioSetting AudioSetting `json:"audio_setting"`
}

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
	Status string `json:"status"`
	FileID int64  `json:"file_id"`
}

type RetrieveResponse struct {
	ResponseWrapper
	File struct {
		DownloadURL string `json:"download_url"`
	} `json:"file"`
}

func initiateTask(text string) (int64, error) {
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
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", BaseURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var t2aResp T2AResponse
	json.NewDecoder(resp.Body).Decode(&t2aResp)
	if t2aResp.BaseResp.StatusCode != 0 {
		return 0, fmt.Errorf("API Error: %s", t2aResp.BaseResp.StatusMsg)
	}
	return t2aResp.TaskID, nil
}

func queryStatus(taskID int64) (string, int64, error) {
	url := fmt.Sprintf(QueryURL, fmt.Sprintf("%d", taskID))
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var queryResp QueryResponse
	json.NewDecoder(resp.Body).Decode(&queryResp)
	return queryResp.Status, queryResp.FileID, nil
}

func getDownloadURL(fileID int64) (string, error) {
	url := fmt.Sprintf(RetrieveURL, fmt.Sprintf("%d", fileID))
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var retrieveResp RetrieveResponse
	json.NewDecoder(resp.Body).Decode(&retrieveResp)
	return retrieveResp.File.DownloadURL, nil
}

func downloadAndExtract(url string) ([]byte, error) {
	// 简单起见，这里先只下载，解压逻辑比较复杂，建议复用之前的 tar 解压代码
	// 但为了保持 service 层的纯净，这里应该返回 mp3 的 byte content
	// 由于 tar 解压比较繁琐，我建议先直接返回 tar 的 bytes，或者在这里做解压
	// 考虑到之前的 main.go 已经实现了 extractTar，我们这里应该引入 archive/tar

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取 tar 数据
	// ... (此处需要实现内存解压，或者先存临时文件)
	// 为了快速实现，我们假设这里返回的是 mp3 数据 (如果 API 支持直接下载 mp3 最好，但 MiniMax 是 tar)
	// 暂时返回 nil, fmt.Errorf("Not implemented yet")
	// 实际项目中需要把 main.go 的 extractTar 逻辑搬过来
	return nil, fmt.Errorf("tar extraction not implemented in this snippet")
}

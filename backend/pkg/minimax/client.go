package minimax

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

	"voicepaper/config"
)

// GenerateSpeech 是对外的统一接口，处理所有异步轮询逻辑，直接返回音频二进制数据
func GenerateSpeech(text string) ([]byte, error) {
	cfg := config.GetConfig()

	// 1. 发起请求
	taskID, err := initiateTask(text, cfg)
	if err != nil {
		return nil, err
	}

	// 2. 轮询状态
	var fileID int64
	for {
		status, fid, err := queryStatus(taskID, cfg)
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
	downloadURL, err := getDownloadURL(fileID, cfg)
	if err != nil {
		return nil, err
	}

	// 4. 下载并返回数据
	// MiniMax 返回的是 tar 包，需要下载 -> 解压 -> 提取 mp3
	return downloadAndExtract(downloadURL, cfg)
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

func initiateTask(text string, cfg *config.Config) (int64, error) {
	reqBody := T2ARequest{
		Model: cfg.TTS.Model,
		Text:  text,
		VoiceSetting: VoiceSetting{
			VoiceID: cfg.TTS.VoiceID,
			Speed:   cfg.TTS.Speed,
			Vol:     cfg.TTS.Volume,
			Pitch:   cfg.TTS.Pitch,
		},
		AudioSetting: AudioSetting{
			AudioSampleRate: cfg.TTS.AudioSampleRate,
			Bitrate:         cfg.TTS.Bitrate,
			Format:          cfg.TTS.Format,
			Channel:         cfg.TTS.Channel,
		},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", cfg.MiniMax.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.MiniMax.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Duration(cfg.Network.Timeout) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %w", err)
	}

	var t2aResp T2AResponse
	if err := json.Unmarshal(bodyBytes, &t2aResp); err != nil {
		return 0, fmt.Errorf("解析响应失败: %w, body: %s", err, string(bodyBytes))
	}

	if t2aResp.BaseResp.StatusCode != 0 {
		return 0, fmt.Errorf("API Error: %s (Code: %d)", t2aResp.BaseResp.StatusMsg, t2aResp.BaseResp.StatusCode)
	}
	return t2aResp.TaskID, nil
}

func queryStatus(taskID int64, cfg *config.Config) (string, int64, error) {
	url := fmt.Sprintf(cfg.MiniMax.QueryURL, fmt.Sprintf("%d", taskID))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.MiniMax.APIKey)

	client := &http.Client{
		Timeout: time.Duration(cfg.Network.Timeout) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("读取响应失败: %w", err)
	}

	var queryResp QueryResponse
	if err := json.Unmarshal(bodyBytes, &queryResp); err != nil {
		return "", 0, fmt.Errorf("解析响应失败: %w, body: %s", err, string(bodyBytes))
	}

	if queryResp.BaseResp.StatusCode != 0 {
		return "", 0, fmt.Errorf("API Error: %s (Code: %d)", queryResp.BaseResp.StatusMsg, queryResp.BaseResp.StatusCode)
	}

	return queryResp.Status, queryResp.FileID, nil
}

func getDownloadURL(fileID int64, cfg *config.Config) (string, error) {
	url := fmt.Sprintf(cfg.MiniMax.RetrieveURL, fmt.Sprintf("%d", fileID))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.MiniMax.APIKey)

	client := &http.Client{
		Timeout: time.Duration(cfg.Network.Timeout) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var retrieveResp RetrieveResponse
	if err := json.Unmarshal(bodyBytes, &retrieveResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w, body: %s", err, string(bodyBytes))
	}

	if retrieveResp.BaseResp.StatusCode != 0 {
		return "", fmt.Errorf("API Error: %s (Code: %d)", retrieveResp.BaseResp.StatusMsg, retrieveResp.BaseResp.StatusCode)
	}

	return retrieveResp.File.DownloadURL, nil
}

func downloadAndExtract(url string, cfg *config.Config) ([]byte, error) {
	// 1. 下载 tar 文件到临时目录
	tempDir := cfg.Storage.TempDir
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	tarPath := filepath.Join(tempDir, fmt.Sprintf("audio_%d.tar", time.Now().Unix()))
	defer os.Remove(tarPath) // 清理临时文件

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(tarPath)
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}
	out.Close()

	// 2. 解压 tar 并提取 mp3
	mp3Data, err := extractMP3FromTar(tarPath)
	if err != nil {
		return nil, fmt.Errorf("解压文件失败: %w", err)
	}

	return mp3Data, nil
}

// extractMP3FromTar 从 tar 文件中提取 mp3 数据
func extractMP3FromTar(tarPath string) ([]byte, error) {
	file, err := os.Open(tarPath)
	if err != nil {
		return nil, fmt.Errorf("打开 tar 文件失败: %w", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)
	var mp3Data []byte

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取 tar 条目失败: %w", err)
		}

		// 只处理普通文件，且是 mp3 文件
		if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, ".mp3") {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("读取 mp3 数据失败: %w", err)
			}
			mp3Data = data
			break // 找到第一个 mp3 文件就返回
		}
	}

	if len(mp3Data) == 0 {
		return nil, fmt.Errorf("tar 文件中未找到 mp3 文件")
	}

	return mp3Data, nil
}

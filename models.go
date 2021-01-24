package thumbnaild

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type thumbnailhelper struct {
	ffmpegPath       string
	uploadImageURL   string
	youtubeThumbnail string

	youtuberegex *regexp.Regexp
}

func (m *thumbnailhelper) GetThumbnailVideo(videoURL string) (string, error) {
	if m.youtuberegex.MatchString(videoURL) {
		return m.GetThumbnailYoutube(videoURL)
	}
	fileName := fmt.Sprintf("img_%d.png", time.Now().Unix())
	cmd := exec.Command(m.ffmpegPath, "-i", videoURL, "-ss", "00:00:01.000", "-vframes", "1", fileName)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	if err := cmd.Run(); err != nil {
		return "", err
	}
	defer os.Remove(fileName)
	return m.uploadImage(fileName)

}

func (m *thumbnailhelper) GetThumbnailYoutube(videoURL string) (string, error) {
	if !m.youtuberegex.MatchString(videoURL) {
		return "", errors.New("Not type youtube URL")
	}
	var youtubeID string
	if strings.Contains(videoURL, "/embed/") {
		youtubeID = strings.Split(videoURL, "/embed/")[1]
	} else {
		queryParameters := strings.Split(videoURL, "?v=")[1]
		youtubeID = strings.Split(queryParameters, "&")[0]
	}
	return fmt.Sprintf("https://i.ytimg.com/vi/%s/%d.jpg", youtubeID, 0), nil
}

func NewThumbailHelper(ffmpegPath, uploadImageURL string) *thumbnailhelper {
	if ffmpegPath == "" {
		ffmpegPath = "/usr/bin/ffmpeg"
	}
	if uploadImageURL == "" {
		uploadImageURL = "https://upload.photocloud.mobilelab.vn/upload"
	}
	regexYoutubeLink, _ := regexp.Compile("^.*((youtu.be\\/)|(v\\/)|(\\/u\\/\\w\\/)|(embed\\/)|(watch\\?))\\??v?=?([^#&?]*).*")
	return &thumbnailhelper{
		ffmpegPath:     ffmpegPath,
		uploadImageURL: uploadImageURL,
		youtuberegex:   regexYoutubeLink,
	}
}

func (m *thumbnailhelper) uploadImage(imageURL string) (string, error) {
	extraParams := map[string]string{
		"name":  imageURL,
		"email": "minhnv@sonek.vn",
	}
	fileRequest, err := m.newfileUploadRequest(m.uploadImageURL, extraParams, "files", imageURL)
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	responses, err := client.Do(fileRequest)
	if err != nil {
		return "", err
	}
	defer responses.Body.Close()
	bodybytes, err := ioutil.ReadAll(responses.Body)
	if err != nil {
		return "", err
	}
	var jsonResponse map[string]interface{}
	err = json.Unmarshal(bodybytes, &jsonResponse)
	if err != nil {
		return "", err
	}
	if jsonResponse["uploaded_files"] == nil {
		return "", errors.New("Bad responses")
	}
	listUrl := jsonResponse["uploaded_files"].([]interface{})
	if listUrl == nil || len(listUrl) == 0 {
		return "", errors.New("Bad responses")
	}
	return listUrl[0].(string), nil
}

func (m *thumbnailhelper) newfileUploadRequest(uri string, params map[string]string, paramName, fileName string) (*http.Request, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(fileName))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

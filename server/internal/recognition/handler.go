package recognition

import (
	"encoding/binary"
	"fmt"
	"go-shazam/internal/audio"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type RecognitionHandler struct {
	service *RecognitionService
}

func NewRecognitionHandler(service *RecognitionService) *RecognitionHandler {
	return &RecognitionHandler{service: service}
}

func RegisterRoutes(r *gin.Engine, h *RecognitionHandler) {
	r.GET("/api/recognize/ws", h.HandleWebSocket)
}

func (h *RecognitionHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var audioData []float64
	sampleRate := 44100 // Default

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if messageType == websocket.BinaryMessage {
			// Append audio chunks
			floats := bytesToFloats(p)
			audioData = append(audioData, floats...)
		} else if messageType == websocket.TextMessage {
			msg := string(p)
			if strings.HasPrefix(msg, "start") {
				// Reset and optional set sample rate: "start:48000"
				audioData = []float64{}
				parts := strings.Split(msg, ":")
				if len(parts) > 1 {
					if rate, err := strconv.Atoi(parts[1]); err == nil {
						sampleRate = rate
					}
				}
			} else if msg == "stop" || msg == "analyze" {
				if len(audioData) == 0 {
					conn.WriteJSON(gin.H{"error": "no audio data received"})
					continue
				}

				// Resample to match the database sample rate
				if sampleRate != audio.TargetSampleRate {
					var err error
					audioData, err = audio.Resample(audioData, sampleRate, audio.TargetSampleRate)
					if err != nil {
						conn.WriteJSON(gin.H{"error": fmt.Sprintf("resampling error: %s", err.Error())})
						continue
					}
				}

				fragments, err := audio.ProcessAudio(audioData, audio.TargetSampleRate)
				if err != nil {
					conn.WriteJSON(gin.H{"error": fmt.Sprintf("processing error: %s", err.Error())})
					continue
				}

				match, err := h.service.IdentifySong(c.Request.Context(), fragments, audio.TargetSampleRate)
				if err != nil {
					conn.WriteJSON(gin.H{"error": fmt.Sprintf("recognition error: %s", err.Error())})
				} else if match == nil {
					conn.WriteJSON(gin.H{"found": false})
				} else {
					conn.WriteJSON(gin.H{
						"found":       true,
						"song":        match.Song,
						"time_offset": match.TimeOffset,
						"score":       match.Score,
					})
				}

				// Clear buffer after analysis
				// Usually "stop" implies end of this session.
				audioData = []float64{}
			}
		}
	}
}

func bytesToFloats(b []byte) []float64 {
	floats := make([]float64, len(b)/4)
	for i := 0; i < len(floats); i++ {
		bits := binary.LittleEndian.Uint32(b[i*4 : (i+1)*4])
		floats[i] = float64(math.Float32frombits(bits))
	}
	return floats
}

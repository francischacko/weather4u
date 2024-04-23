package models

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Index        int     `json:"index"`
	Delta        Delta   `json:"delta"`
	FinishReason *string `json:"finish_reason,omitempty"`
}

type Delta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	Stream      bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type locn struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

type acq struct {
	Pm float32 `json:"pm2_5"`
}

type currnt struct {
	Realfeel float32 `json:"feelslike_c"`
	Humidity int     `json:"humidity"`
	Air      acq     `json:"air_quality"`
}

type ResponseBody struct {
	Location locn   `json:"location"`
	Current  currnt `json:"current"`
}

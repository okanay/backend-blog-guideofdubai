package types

// TranslateRequest, çeviri isteği için gerekli verileri içerir
type TranslateRequest struct {
	HTML           string `json:"html" binding:"required"`
	SourceLanguage string `json:"sourceLanguage" binding:"required"`
	TargetLanguage string `json:"targetLanguage" binding:"required"`
	BlogID         string `json:"blogId"` // İsteğe bağlı, belirli bir blogu çevirmek için
}

// TranslateResponse, çeviri yanıtı için yapı
type TranslateResponse struct {
	TranslatedHTML string `json:"translatedHTML"`
	SourceLanguage string `json:"sourceLanguage"`
	TargetLanguage string `json:"targetLanguage"`
}

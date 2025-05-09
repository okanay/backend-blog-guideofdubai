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

type GenerateMetadataRequest struct {
	HTML     string `json:"html" binding:"required"`
	Language string `json:"language" binding:"required"`
}

// GenerateMetadataResponse, AI tarafından oluşturulan metadata yanıtını içerir
type GenerateMetadataResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	// Categories  []CategoryView `json:"categories"`
	// Tags        []TagView      `json:"tags"`
}

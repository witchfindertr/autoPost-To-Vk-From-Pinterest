package models

const PIN_FIELDS = "id,link,note,url,attribution,color,board,counts,created_at,creator,image,media,metadata,original_link"

type PinterestIni struct {
	Token     string
	BoardPic string
	BoardVid string
	Iter      int8
	PinUser   string
}

type Pinterest struct {
	Data []struct {
		Attribution interface{} `json:"attribution"`
		Creator     struct {
			URL       string `json:"url"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			ID        string `json:"id"`
		} `json:"creator"`
		URL   string `json:"url"`
		Media struct {
			Type string `json:"type"`
		} `json:"media"`
		CreatedAt    string `json:"created_at"`
		OriginalLink string `json:"original_link"`
		Note         string `json:"note"`
		Color        string `json:"color"`
		Link         string `json:"link"`
		Board        struct {
			URL  string `json:"url"`
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"board"`
		Image struct {
			Original struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"original"`
		} `json:"image"`
		Counts struct {
			Saves    int `json:"saves"`
			Comments int `json:"comments"`
		} `json:"counts"`
		ID       string `json:"id"`
		Metadata struct {
			Article struct {
				PublishedAt interface{}   `json:"published_at"`
				Description string        `json:"description"`
				Name        string        `json:"name"`
				Authors     []interface{} `json:"authors"`
			} `json:"article"`
			Link struct {
				Locale      string `json:"locale"`
				Title       string `json:"title"`
				SiteName    string `json:"site_name"`
				Description string `json:"description"`
				Favicon     string `json:"favicon"`
			} `json:"link"`
		} `json:"metadata"`
	} `json:"data"`
	Page struct {
		Cursor string `json:"cursor"`
		Next   string `json:"next"`
	} `json:"page"`
}

//type Pin struct {
//	Id           string       `json:"id"`
//	Link         string       `json:"link"`
//	Url          string       `json:"url"`
//	//Creator      Creator      `json:"creator"`
//	//Board        Board        `json:"board"`
//	CreatedAt    iso8601.Time `json:"created_at"`
//	Note         string       `json:"note"`
//	Color        string       `json:"color"`
//	Counts       PinCounts    `json:"counts"`
//	Media        Media        `json:"json:media"`
//	OriginalLink string       `json:"original_link"`
//	Attribution  Attribution  `json:"attribution"`
//	Image        PinImage     `json:"image"`
//	Metadata     PinMetadata  `json:"metadata"`
//}

//type PinImage struct {
//	Original Image `json:"original"`
//}

//type PinCounts struct {
//	Likes    int32 `json:"likes"`
//	Comments int32 `json:"comments"`
//	Repins   int32 `json:"repins"`
//}

//type Media struct {
//	Type string `json:"type"`
//}

//type Attribution struct {
//	Title              string `json:"title"`
//	Url                string `json:"url"`
//	ProviderIconUrl    string `json:"provider_icon_url"`
//	AuthorName         string `json:"author_name"`
//	ProviderFaviconUrl string `json:"provider_favicon_url"`
//	AuthorUrl          string `json:"author_url"`
//	ProviderName       string `json:"provider_name"`
//}

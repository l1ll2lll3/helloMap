package main

type RouteChangeData struct {
	Meta struct {
		SendEventID      string `json:"sendEventID"`
		PageLocation     string `json:"pageLocation"`
		Host             string `json:"host"`
		Path             string `json:"path"`
		Query            string `json:"query"`
		Protocol         string `json:"protocol"`
		PageTitle        string `json:"pageTitle"`
		PCode            int    `json:"pCode"`
		ProjectAccessKey string `json:"projectAccessKey"`
		Screen           struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"screen"`
		SessionID string `json:"sessionID"`
		UserAgent string `json:"userAgent"`
		UserID    string `json:"userID"`
	} `json:"meta"`
	RouteChange []struct {
		RouterChangeTiming struct {
			IsComplete     bool   `json:"isComplete"`
			StartTimeStamp int64  `json:"startTimeStamp"`
			EndTimeStamp   int64  `json:"endTimeStamp"`
			LoadTime       int    `json:"loadTime"`
			PageLocation   string `json:"pageLocation"`
		} `json:"routerChangeTiming"`
		Resource []struct {
			StartTime      int    `json:"startTime"`
			StartTimeStamp int64  `json:"startTimeStamp"`
			EventID        string `json:"eventID"`
			Type           string `json:"type"`
			URL            string `json:"url"`
			URLHost        string `json:"urlHost"`
			URLPath        string `json:"urlPath"`
			URLQuery       string `json:"urlQuery"`
			URLProtocol    string `json:"urlProtocol"`
			Timing         struct {
				Redirect struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"redirect"`
				Cache struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"cache"`
				Connect struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"connect"`
				DNS struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"dns"`
				Ssl struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"ssl"`
				Download struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"download"`
				FirstByte struct {
					Duration int `json:"duration"`
					Start    int `json:"start"`
				} `json:"firstByte"`
				Duration int `json:"duration"`
				Size     int `json:"size"`
			} `json:"timing"`
			ResourceInfo struct {
				Method string `json:"method"`
				Status int    `json:"status"`
			} `json:"resourceInfo"`
			TraceInfo struct {
				MtID string `json:"mtID"`
				TxID string `json:"txID"`
			} `json:"traceInfo"`
		} `json:"resource"`
	} `json:"routeChange"`
}

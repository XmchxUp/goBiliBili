package payload

type BiliBiliDynamicSimplifyResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		HasMore bool   `json:"has_more"`
		Offset  string `json:"offset"`
	}
}

type BiliBiliDynamicResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		HasMore bool `json:"has_more"`
		Items   []struct {
			Modules struct {
				ModuleAuthor struct {
					Avatar struct {
						ContainerSize struct {
							Height float64 `json:"height"`
							Width  float64 `json:"width"`
						} `json:"container_size"`
					} `json:"avatar"`
				} `json:"module_author"`
				ModuleDynamic struct {
					Major struct {
						Archive struct {
							Cover   string `json:"cover"`
							Desc    string `json:"desc"`
							JumpURL string `json:"jump_url"`
							Stat    struct {
								Danmaku string `json:"danmaku"`
								Play    string `json:"play"`
							} `json:"stat"`
							Pics []struct {
								Height uint64  `json:"height"`
								Width  uint64  `json:"width"`
								Size   float64 `json:"size"`
								URL    string  `json:"url"`
							} `json:"pics"`
						} `json:"archive"`
					} `json:"major"`
				} `json:"module_dynamic"`
			} `json:"modules"`
			Type string `json:"type"`
		} `json:"items"`
		Offset string `json:"offset"`
	} `json:"data"`
}

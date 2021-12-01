package sdsshared

//DataResource is the interface each Resource service uses and is a central library unit used for a centralised server facility that handles JWT checking centrally.
type DataResource interface {
	Startup() error
	UpdateDataset(*VersionManager) (*VersionManager, error)
	//Retrieve takes query token and map[string]string group of query args
	// received in the GET request.
	// Returned []byte is JSON representation of SimpleData
	Retrieve(string, map[string]string) (SimpleData, error)
	Shutdown() error
}

//SimpleData is the standard response interface to the end user. Data resources should return
type SimpleData struct {
	ResultCount    int               `json:"result_count"`
	RequestOptions map[string]string `json:"request_options,omitempty"`
	Meta           struct {
		Resource    string   `json:"resource"`
		LastUpdated string   `json:"dataset_updated"` //time.RFC3339
		DataSources []string `json:"data_sources"`
	} `json:"meta"`
	Data struct {
		//... Requested data schema...
		Values interface{} `json:"values"`
	} `json:"data,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

//VersionManager is the struct that allows the instance to check its current used
// dataset version and compare it against the latest available to judge update need
type VersionManager struct {
	CurrentVersion int `json:"version,omitempty"`
	//Repo is where to get the dataset archive from (URL)
	Repo string `json:"repo,omitempty"`
	//LastUpdated is when the Repo was last updated from latest version (time.RFC3339)
	LastUpdated string `json:"dataset_updated"`
	//List of initial data sources gained from last update from repo
	DataSources []string `json:"data_sources"`
}

func (vt *VersionManager) UpdateDataset(dr DataResource) error {
	var err error
	vt, err = dr.UpdateDataset(vt)
	if err != nil {
		return err
	}

	return nil
}

//DataResourceImplementorTemplate is a simple outline of the basic structure that can
// implement the full DataResource interface. See `badgerdb` for best practise
type DataResourceImplementorTemplate struct {
	DatabaseHandle interface{}
	versioner      VersionManager
}

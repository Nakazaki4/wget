package downloader

type Config struct {
	Background   bool // background download
	OutputName   string
	OutputPath   string // where the file will be saved
	InputFile    string // files with links to be downlaoded
	RateLimit    string
	Mirror       bool
	Reject       string // reject file suffixes
	Exclude      string // exclude path
	ConvertLinks bool   // convert links for offline viewing
	URLs         []string
}
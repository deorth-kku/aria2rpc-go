package aria2rpc

// Status represents aria2 task status.
type Status struct {
	GID                    string          `json:"gid,omitempty"`
	Status                 string          `json:"status,omitempty"`
	TotalLength            string          `json:"totalLength,omitempty"`
	CompletedLength        string          `json:"completedLength,omitempty"`
	UploadLength           string          `json:"uploadLength,omitempty"`
	Bitfield               string          `json:"bitfield,omitempty"`
	DownloadSpeed          string          `json:"downloadSpeed,omitempty"`
	UploadSpeed            string          `json:"uploadSpeed,omitempty"`
	InfoHash               string          `json:"infoHash,omitempty"`
	NumSeeders             string          `json:"numSeeders,omitempty"`
	Seeder                 string          `json:"seeder,omitempty"`
	PieceLength            string          `json:"pieceLength,omitempty"`
	NumPieces              string          `json:"numPieces,omitempty"`
	Connections            string          `json:"connections,omitempty"`
	ErrorCode              string          `json:"errorCode,omitempty"`
	ErrorMessage           string          `json:"errorMessage,omitempty"`
	FollowedBy             []string        `json:"followedBy,omitempty"`
	BelongsTo              string          `json:"belongsTo,omitempty"`
	Dir                    string          `json:"dir,omitempty"`
	Files                  []FileInfo      `json:"files,omitempty"`
	Bittorrent             *BittorrentInfo `json:"bittorrent,omitempty"`
	VerifiedLength         string          `json:"verifiedLength,omitempty"`
	VerifyIntegrityPending string          `json:"verifyIntegrityPending,omitempty"`
}

type URIStatus struct {
	URI    string `json:"uri,omitempty"`
	Status string `json:"status,omitempty"`
}

type FileInfo struct {
	Index           string      `json:"index,omitempty"`
	Path            string      `json:"path,omitempty"`
	Length          string      `json:"length,omitempty"`
	CompletedLength string      `json:"completedLength,omitempty"`
	Selected        string      `json:"selected,omitempty"`
	URIs            []URIStatus `json:"uris,omitempty"`
}

type PeerInfo struct {
	PeerID        string `json:"peerId,omitempty"`
	IP            string `json:"ip,omitempty"`
	Port          string `json:"port,omitempty"`
	Bitfield      string `json:"bitfield,omitempty"`
	AmChoking     string `json:"amChoking,omitempty"`
	PeerChoking   string `json:"peerChoking,omitempty"`
	DownloadSpeed string `json:"downloadSpeed,omitempty"`
	UploadSpeed   string `json:"uploadSpeed,omitempty"`
	Seeder        string `json:"seeder,omitempty"`
}

type ServerInfo struct {
	Index   string      `json:"index,omitempty"`
	Servers []SubServer `json:"servers,omitempty"`
}

type SubServer struct {
	URI           string `json:"uri,omitempty"`
	CurrentURI    string `json:"currentUri,omitempty"`
	DownloadSpeed string `json:"downloadSpeed,omitempty"`
}

type BittorrentInfo struct {
	AnnounceList [][]string        `json:"announceList,omitempty"`
	Comment      string            `json:"comment,omitempty"`
	CreationDate int64             `json:"creationDate,omitempty"`
	Mode         string            `json:"mode,omitempty"`
	Info         map[string]string `json:"info,omitempty"`
}

type GlobalStat struct {
	DownloadSpeed   string `json:"downloadSpeed,omitempty"`
	UploadSpeed     string `json:"uploadSpeed,omitempty"`
	NumActive       string `json:"numActive,omitempty"`
	NumWaiting      string `json:"numWaiting,omitempty"`
	NumStopped      string `json:"numStopped,omitempty"`
	NumStoppedTotal string `json:"numStoppedTotal,omitempty"`
}

type VersionInfo struct {
	Version         string   `json:"version,omitempty"`
	EnabledFeatures []string `json:"enabledFeatures,omitempty"`
}

type SessionInfo struct {
	SessionID string `json:"sessionId,omitempty"`
}

// DownloadEvent is payload for aria2.onDownload* notifications.
type DownloadEvent struct {
	GID string `json:"gid,omitempty"`
}

// Multicall item for system.multicall.
type Multicall struct {
	MethodName string
	Params     []any
}

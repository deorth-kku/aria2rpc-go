package aria2rpc

import (
	"time"
)

// DownloadStatus represents the status of a download.
type DownloadStatus string

const (
	// StatusActive represents currently downloading/seeding downloads
	StatusActive DownloadStatus = "active"
	// StatusWaiting represents downloads in the queue
	StatusWaiting DownloadStatus = "waiting"
	// StatusPaused represents paused downloads
	StatusPaused DownloadStatus = "paused"
	// StatusError represents downloads that were stopped because of error
	StatusError DownloadStatus = "error"
	// StatusCompleted represents stopped and completed downloads
	StatusCompleted DownloadStatus = "complete"
	// StatusRemoved represents the downloads removed by user
	StatusRemoved DownloadStatus = "removed"
)

// ExitStatus is an integer returned by aria2 for downloads which describes why a download exited.
type ExitStatus uint8

const (
	// Success indicates that all downloads were successful.
	Success ExitStatus = iota

	// UnknownError indicates that an unknown error occurred.
	UnknownError

	// Timeout indicates that a timeout occurred.
	Timeout

	// ResourceNotFound indicates that a resource was not found.
	ResourceNotFound

	// ResourceNotFoundReached indicates that aria2 saw the specified number of "resource not found" error.
	ResourceNotFoundReached

	// DownloadSpeedTooSlow indicates that a download aborted because download speed was too slow.
	DownloadSpeedTooSlow

	// NetworkError indicates that a network problem occurred.
	NetworkError

	// UnfinishedDownloads indicates that there were unfinished downloads.
	UnfinishedDownloads

	// RemoteNoResume indicates that the remote server did not support resume when resume was required.
	RemoteNoResume

	// NotEnoughDiskSpace indicates that there was not enough disk space available.
	NotEnoughDiskSpace

	// PieceLengthMismatch indicates that the piece length was different from one in .aria2 control file.
	PieceLengthMismatch

	// SameFileBeingDownloaded indicates that aria2 was downloading same file at that moment.
	SameFileBeingDownloaded

	// SameInfoHashBeingDownloaded indicates that aria2 was downloading same info hash torrent at that moment.
	SameInfoHashBeingDownloaded

	// FileAlreadyExists indicates that the file already existed.
	FileAlreadyExists

	// RenamingFailed indicates that renaming the file failed.
	RenamingFailed

	// CouldNotOpenExistingFile indicates that aria2 could not open existing file.
	CouldNotOpenExistingFile

	// CouldNotCreateNewFile indicates that aria2 could not create new file or truncate existing file.
	CouldNotCreateNewFile

	// FileIOError indicates that a file I/O error occurred.
	FileIOError

	// CouldNotCreateDirectory indicates that aria2 could not create directory.
	CouldNotCreateDirectory

	// NameResolutionFailed indicates that the name resolution failed.
	NameResolutionFailed

	// MetalinkParsingFailed indicates that aria2 could not parse Metalink document.
	MetalinkParsingFailed

	// FTPCommandFailed indicates that the FTP command failed.
	FTPCommandFailed

	// HTTPResponseHeaderBad indicates that the HTTP response header was bad or unexpected.
	HTTPResponseHeaderBad

	// TooManyRedirects indicates that too many redirects occurred.
	TooManyRedirects

	// HTTPAuthorizationFailed indicates that HTTP authorization failed.
	HTTPAuthorizationFailed

	// BencodedFileParseError indicates that aria2 could not parse bencoded file.
	BencodedFileParseError

	// TorrentFileCorrupt indicates that the ".torrent" file was corrupted or missing information.
	TorrentFileCorrupt

	// MagnetURIBad indicates that the magnet URI was bad.
	MagnetURIBad

	// RemoteServerHandleRequestError indicates that the remote server was unable to handle the request.
	RemoteServerHandleRequestError

	// PartialDownload indicates that the download was partially completed.
	PartialDownload

	// Removed indicates that the download was removed by the user.
	Removed
)

// TorrentMode represents the file mode of the torrent
type TorrentMode string

const (
	// TorrentModeSingle represents the file mode single
	TorrentModeSingle TorrentMode = "single"
	// TorrentModeMulti represents the file mode multi
	TorrentModeMulti TorrentMode = "multi"
)

// URIStatusType represents the status of a URI.
type URIStatusType string

const (
	// URIStatusUsed represents the state of the URI being used
	URIStatusUsed URIStatusType = "used"
	// URIStatusWaiting represents the state of the URI waiting in the queue
	URIStatusWaiting URIStatusType = "waiting"
)

// Status represents aria2 task status.
type Status struct {
	GID                    string            `json:"gid,omitzero"`
	Status                 DownloadStatus    `json:"status,omitzero"`
	TotalLength            uint              `json:"totalLength,string,omitzero"`
	CompletedLength        uint              `json:"completedLength,string,omitzero"`
	UploadLength           uint              `json:"uploadLength,string,omitzero"`
	BitField               string            `json:"bitfield,omitzero"`
	DownloadSpeed          uint              `json:"downloadSpeed,string,omitzero"`
	UploadSpeed            uint              `json:"uploadSpeed,string,omitzero"`
	InfoHash               string            `json:"infoHash,omitzero"`
	NumSeeders             uint              `json:"numSeeders,string,omitzero"`
	Seeder                 bool              `json:"seeder,string,omitzero"`
	PieceLength            uint              `json:"pieceLength,string,omitzero"`
	NumPieces              uint              `json:"numPieces,string,omitzero"`
	Connections            uint              `json:"connections,string,omitzero"`
	ErrorCode              ExitStatus        `json:"errorCode,string,omitzero"`
	ErrorMessage           string            `json:"errorMessage,omitzero"`
	FollowedBy             []string          `json:"followedBy,omitzero"`
	Following              string            `json:"following,omitzero"`
	BelongsTo              string            `json:"belongsTo,omitzero"`
	Dir                    string            `json:"dir,omitzero"`
	Files                  []File            `json:"files,omitzero"`
	BitTorrent             *BitTorrentStatus `json:"bittorrent,omitzero"`
	VerifiedLength         uint              `json:"verifiedLength,string,omitzero"`
	VerifyIntegrityPending bool              `json:"verifyIntegrityPending,string,omitzero"`
}

type URIStatus struct {
	URI    string        `json:"uri,omitzero"`
	Status URIStatusType `json:"status,omitzero"`
}

type File struct {
	Index           uint   `json:"index,string,omitzero"`
	Path            string `json:"path,omitzero"`
	Length          uint   `json:"length,string,omitzero"`
	CompletedLength uint   `json:"completedLength,string,omitzero"`
	Selected        bool   `json:"selected,string,omitzero"`
	URIs            []URI  `json:"uris,omitzero"`
}

// URI represents a URI used in a download.
type URI struct {
	URI    string        `json:"uri,omitzero"`
	Status URIStatusType `json:"status,omitzero"`
}

type PeerInfo struct {
	PeerID        string `json:"peerId,omitzero"`
	IP            string `json:"ip,omitzero"`
	Bitfield      string `json:"bitfield,omitzero"`
	DownloadSpeed uint   `json:"downloadSpeed,string,omitzero"`
	UploadSpeed   uint   `json:"uploadSpeed,string,omitzero"`
	// struct fields align to save on ram
	Port        uint16 `json:"port,string,omitzero"`
	AmChoking   bool   `json:"amChoking,string,omitzero"`
	PeerChoking bool   `json:"peerChoking,string,omitzero"`
	Seeder      bool   `json:"seeder,string,omitzero"`
}

type ServerInfo struct {
	Index   uint        `json:"index,string,omitzero"`
	Servers []SubServer `json:"servers,omitzero"`
}

type SubServer struct {
	URI           string `json:"uri,omitzero"`
	CurrentURI    string `json:"currentUri,omitzero"`
	DownloadSpeed uint   `json:"downloadSpeed,string,omitzero"`
}

// BitTorrentStatus holds information for a BitTorrent download.
type BitTorrentStatus struct {
	AnnounceList [][]string           `json:"announceList,omitzero"`
	Comment      string               `json:"comment,omitzero"`
	CreationDate time.Time            `json:"creationDate,omitzero"`
	Mode         TorrentMode          `json:"mode,omitzero"`
	Info         BitTorrentStatusInfo `json:"info,omitzero"`
}

// BitTorrentStatusInfo holds information from the info dictionary.
type BitTorrentStatusInfo struct {
	Name string `json:"name,omitzero"`
}

type GlobalStat struct {
	DownloadSpeed   uint `json:"downloadSpeed,string,omitzero"`
	UploadSpeed     uint `json:"uploadSpeed,string,omitzero"`
	NumActive       uint `json:"numActive,string,omitzero"`
	NumWaiting      uint `json:"numWaiting,string,omitzero"`
	NumStopped      uint `json:"numStopped,string,omitzero"`
	NumStoppedTotal uint `json:"numStoppedTotal,string,omitzero"`
}

type VersionInfo struct {
	Version         string   `json:"version,omitzero"`
	EnabledFeatures []string `json:"enabledFeatures,omitzero"`
}

type SessionInfo struct {
	SessionID string `json:"sessionId,omitzero"`
}

// DownloadEvent is payload for aria2.onDownload* notifications.
type DownloadEvent struct {
	GID string `json:"gid,omitzero"`
}

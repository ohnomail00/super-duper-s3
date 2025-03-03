package engine

// PartPlan represents the plan for a single file part.
// Index: the part's index.
// Server: the server where the part is stored.
// Offset: the offset in the original file.
// Length: the length of the part.
type PartPlan struct {
	Index  int
	Server Server
	Offset int64
	Length int64
}

// FileUploadPlan contains the file parts, ordered by Index.
type FileUploadPlan struct {
	Parts []PartPlan
}

package rao

type WorkerConfig struct {
	WorkerName         string
	AddPieceMax        int
	PreCommit1Max      int
	PreCommit2Max      int
	Commit2Max         int
	DiskHoldMax        int
	APDiskHoldMax      int
	ForceP1FromLocalAP bool
	ForceP2FromLocalP1 bool
	ForceC2FromLocalP2 bool
	IsPlanOffline      bool
	AllowP2C2Parallel  bool
	IgnoreOutOfSpace   bool
}

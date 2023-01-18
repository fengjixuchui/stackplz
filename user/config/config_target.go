package config

type TargetConfig struct {
    Name             string
    Uid              uint64
    Pid              uint64
    TidBlacklist     [MAX_TID_BLACKLIST_COUNT]uint32
    TidBlacklistMask uint32
    LibraryDirs      []string
    DataDir          string
    Abi              string
}

func NewTargetConfig() *TargetConfig {
    config := &TargetConfig{}
    return config
}

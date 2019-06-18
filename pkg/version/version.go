package version

type MajorVersion struct {
	Major uint32
}

type MinorVersion struct {
	Major uint32
	Minor uint32
}

type PatchVersion struct {
	Major uint32
	Minor uint32
	Patch uint32
}

type Version struct {
	Major  uint32
	Minor  uint32
	Patch  uint32
	Prefix string
	Suffix string
}

func (v *PatchVersion) IsEqual(other *PatchVersion) bool {
	return *v == *other
}
func (v *Version) IsEqual(other *Version) bool {
	return *v == *other
}

package artifact

type SerializedArtifact struct {
    Type ArtifactType `json:"type"`
    Image int `json:"image"`
    Name string `json:"name"`
    Cost int `json:"cost"`
    Powers []Power `json:"powers"`
    Requirements []Requirement `json:"requirements"`
}

func SerializeArtifact(artifact *Artifact) SerializedArtifact {
    return SerializedArtifact{
        Type: artifact.Type,
        Image: artifact.Image,
        Name: artifact.Name,
        Cost: artifact.Cost,
        Powers: append(make([]Power, 0), artifact.Powers...),
        Requirements: append(make([]Requirement, 0), artifact.Requirements...),
    }
}

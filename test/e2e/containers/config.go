package containers

// ImageConfig contains all images and their respective tags
// needed for running e2e tests.
type ImageConfig struct {
	BabylonRepository string
	BabylonTag        string

	RelayerRepository string
	RelayerTag        string
}

//nolint:deadcode
const (
	// name of babylon container produced by running `make localnet-build-env`
	babylonContainerName = "babylonchain/babylond"
	babylonContainerTag  = "latest"

	hermesRelayerRepository = "informalsystems/hermes"
	hermesRelayerTag        = "v1.8.2"
	// Built using the `build-cosmos-relayer-docker` target on an Intel (amd64) machine and pushed to ECR
	cosmosRelayerRepository = "public.ecr.aws/t9e9i3h0/cosmos-relayer"
	// TODO: Replace with version tag once we have a working version
	cosmosRelayerTag = "main"
)

// NewImageConfig returns ImageConfig needed for running e2e test.
// If isUpgrade is true, returns images for running the upgrade
// If isFork is true, utilizes provided fork height to initiate fork logic
func NewImageConfig(isCosmosRelayer bool) ImageConfig {
	config := ImageConfig{}

	// set relayer image name / tag
	if isCosmosRelayer {
		config.RelayerRepository = cosmosRelayerRepository
		config.RelayerTag = cosmosRelayerTag
	} else {
		config.RelayerRepository = hermesRelayerRepository
		config.RelayerTag = hermesRelayerTag
	}

	// set Babylon image name / tag
	config.BabylonRepository = babylonContainerName
	config.BabylonTag = babylonContainerTag

	return config
}

// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/loggo"
	gitjujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "launchpad.net/gocheck"

	"github.com/juju/juju/cmd/envcmd"
	"github.com/juju/juju/constraints"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/environs/bootstrap"
	"github.com/juju/juju/environs/config"
	"github.com/juju/juju/environs/configstore"
	"github.com/juju/juju/environs/filestorage"
	"github.com/juju/juju/environs/imagemetadata"
	imtesting "github.com/juju/juju/environs/imagemetadata/testing"
	"github.com/juju/juju/environs/simplestreams"
	"github.com/juju/juju/environs/storage"
	"github.com/juju/juju/environs/sync"
	envtesting "github.com/juju/juju/environs/testing"
	envtools "github.com/juju/juju/environs/tools"
	toolstesting "github.com/juju/juju/environs/tools/testing"
	"github.com/juju/juju/instance"
	"github.com/juju/juju/juju/arch"
	"github.com/juju/juju/network"
	"github.com/juju/juju/provider/dummy"
	coretesting "github.com/juju/juju/testing"
	coretools "github.com/juju/juju/tools"
	"github.com/juju/juju/version"
)

type BootstrapSuite struct {
	coretesting.FakeJujuHomeSuite
	gitjujutesting.MgoSuite
	envtesting.ToolsFixture
}

var _ = gc.Suite(&BootstrapSuite{})

func (s *BootstrapSuite) SetUpSuite(c *gc.C) {
	s.FakeJujuHomeSuite.SetUpSuite(c)
	s.MgoSuite.SetUpSuite(c)
}

func (s *BootstrapSuite) SetUpTest(c *gc.C) {
	s.FakeJujuHomeSuite.SetUpTest(c)
	s.MgoSuite.SetUpTest(c)
	s.ToolsFixture.SetUpTest(c)

	// Set up a local source with tools.
	sourceDir := createToolsSource(c, vAll)
	s.PatchValue(&envtools.DefaultBaseURL, sourceDir)

	s.PatchValue(&envtools.BundleTools, toolstesting.GetMockBundleTools(c))
}

func (s *BootstrapSuite) TearDownSuite(c *gc.C) {
	s.MgoSuite.TearDownSuite(c)
	s.FakeJujuHomeSuite.TearDownSuite(c)
}

func (s *BootstrapSuite) TearDownTest(c *gc.C) {
	s.ToolsFixture.TearDownTest(c)
	s.MgoSuite.TearDownTest(c)
	s.FakeJujuHomeSuite.TearDownTest(c)
	dummy.Reset()
}

func (s *BootstrapSuite) TestTest(c *gc.C) {
	for i, test := range bootstrapTests {
		c.Logf("\ntest %d: %s", i, test.info)
		test.run(c)
	}
}

type bootstrapTest struct {
	info string
	// binary version string used to set version.Current
	version string
	sync    bool
	args    []string
	err     string
	// binary version string for expected tools; if set, no default tools
	// will be uploaded before running the test.
	upload      string
	constraints constraints.Value
	placement   string
	hostArch    string
}

func (test bootstrapTest) run(c *gc.C) {
	// Create home with dummy provider and remove all
	// of its envtools.
	env := resetJujuHome(c)

	if test.version != "" {
		useVersion := strings.Replace(test.version, "%LTS%", config.LatestLtsSeries(), 1)
		origVersion := version.Current
		version.Current = version.MustParseBinary(useVersion)
		defer func() { version.Current = origVersion }()
	}

	if test.hostArch != "" {
		origVersion := arch.HostArch
		arch.HostArch = func() string {
			return test.hostArch
		}
		defer func() { arch.HostArch = origVersion }()
	}

	if test.upload == "" {
		usefulVersion := version.Current
		usefulVersion.Series = config.PreferredSeries(env.Config())
		envtesting.AssertUploadFakeToolsVersions(c, env.Storage(), usefulVersion)
	}

	// Run command and check for uploads.
	opc, errc := runCommand(nullContext(c), envcmd.Wrap(new(BootstrapCommand)), test.args...)
	// Check for remaining operations/errors.
	if test.err != "" {
		err := <-errc
		stripped := strings.Replace(err.Error(), "\n", "", -1)
		c.Check(stripped, gc.Matches, test.err)
		return
	}
	if !c.Check(<-errc, gc.IsNil) {
		return
	}

	opPutBootstrapVerifyFile := (<-opc).(dummy.OpPutFile)
	c.Check(opPutBootstrapVerifyFile.Env, gc.Equals, "peckham")
	c.Check(opPutBootstrapVerifyFile.FileName, gc.Equals, environs.VerificationFilename)

	opBootstrap := (<-opc).(dummy.OpBootstrap)
	c.Check(opBootstrap.Env, gc.Equals, "peckham")
	c.Check(opBootstrap.Args.Constraints, gc.DeepEquals, test.constraints)
	c.Check(opBootstrap.Args.Placement, gc.Equals, test.placement)

	opFinalizeBootstrap := (<-opc).(dummy.OpFinalizeBootstrap)
	c.Check(opFinalizeBootstrap.Env, gc.Equals, "peckham")
	c.Check(opFinalizeBootstrap.MachineConfig.Tools, gc.NotNil)
	if test.upload != "" {
		c.Check(opFinalizeBootstrap.MachineConfig.Tools.Version.String(), gc.Equals, test.upload)
	}

	store, err := configstore.Default()
	c.Assert(err, gc.IsNil)
	// Check a CA cert/key was generated by reloading the environment.
	env, err = environs.NewFromName("peckham", store)
	c.Assert(err, gc.IsNil)
	_, hasCert := env.Config().CACert()
	c.Check(hasCert, gc.Equals, true)
	_, hasKey := env.Config().CAPrivateKey()
	c.Check(hasKey, gc.Equals, true)
	info, err := store.ReadInfo("peckham")
	c.Assert(err, gc.IsNil)
	c.Assert(info, gc.NotNil)
	c.Assert(info.APIEndpoint().Addresses, gc.DeepEquals, []string{"localhost:17070"})
}

var bootstrapTests = []bootstrapTest{{
	info: "no args, no error, no upload, no constraints",
}, {
	info: "bad --constraints",
	args: []string{"--constraints", "bad=wrong"},
	err:  `invalid value "bad=wrong" for flag --constraints: unknown constraint "bad"`,
}, {
	info: "conflicting --constraints",
	args: []string{"--constraints", "instance-type=foo mem=4G"},
	err:  `ambiguous constraints: "instance-type" overlaps with "mem"`,
}, {
	info: "bad --series",
	args: []string{"--series", "1bad1"},
	err:  `invalid value "1bad1" for flag --series: invalid series name "1bad1"`,
}, {
	info: "lonely --series",
	args: []string{"--series", "fine"},
	err:  `--series requires --upload-tools`,
}, {
	info: "lonely --upload-series",
	args: []string{"--upload-series", "fine"},
	err:  `--upload-series requires --upload-tools`,
}, {
	info: "--upload-series with --series",
	args: []string{"--upload-tools", "--upload-series", "foo", "--series", "bar"},
	err:  `--upload-series and --series can't be used together`,
}, {
	info:    "bad environment",
	version: "1.2.3-%LTS%-amd64",
	args:    []string{"-e", "brokenenv"},
	err:     `failed to bootstrap environment: dummy.Bootstrap is broken`,
}, {
	info:        "constraints",
	args:        []string{"--constraints", "mem=4G cpu-cores=4"},
	constraints: constraints.MustParse("mem=4G cpu-cores=4"),
}, {
	info:        "unsupported constraint passed through but no error",
	args:        []string{"--constraints", "mem=4G cpu-cores=4 cpu-power=10"},
	constraints: constraints.MustParse("mem=4G cpu-cores=4 cpu-power=10"),
}, {
	info:        "--upload-tools uses arch from constraint if it matches current version",
	version:     "1.3.3-saucy-ppc64el",
	hostArch:    "ppc64el",
	args:        []string{"--upload-tools", "--constraints", "arch=ppc64el"},
	upload:      "1.3.3.1-raring-ppc64el", // from version.Current
	constraints: constraints.MustParse("arch=ppc64el"),
}, {
	info:     "--upload-tools rejects mismatched arch",
	version:  "1.3.3-saucy-amd64",
	hostArch: "amd64",
	args:     []string{"--upload-tools", "--constraints", "arch=ppc64el"},
	err:      `failed to bootstrap environment: cannot build tools for "ppc64el" using a machine running on "amd64"`,
}, {
	info:     "--upload-tools rejects non-supported arch",
	version:  "1.3.3-saucy-arm64",
	hostArch: "arm64",
	args:     []string{"--upload-tools"},
	err:      `failed to bootstrap environment: environment "peckham" of type dummy does not support instances running on "arm64"`,
}, {
	info:    "--upload-tools always bumps build number",
	version: "1.2.3.4-raring-amd64",
	args:    []string{"--upload-tools"},
	upload:  "1.2.3.5-raring-amd64",
}, {
	info:      "placement",
	args:      []string{"--to", "something"},
	placement: "something",
}, {
	info: "additional args",
	args: []string{"anything", "else"},
	err:  `unrecognized args: \["anything" "else"\]`,
}}

func (s *BootstrapSuite) TestBootstrapTwice(c *gc.C) {
	env := resetJujuHome(c)
	defaultSeriesVersion := version.Current
	defaultSeriesVersion.Series = config.PreferredSeries(env.Config())
	// Force a dev version by having a non zero build number.
	// This is because we have not uploaded any tools and auto
	// upload is only enabled for dev versions.
	defaultSeriesVersion.Build = 1234
	s.PatchValue(&version.Current, defaultSeriesVersion)

	_, err := coretesting.RunCommand(c, envcmd.Wrap(&BootstrapCommand{}))
	c.Assert(err, gc.IsNil)

	_, err = coretesting.RunCommand(c, envcmd.Wrap(&BootstrapCommand{}))
	c.Assert(err, gc.ErrorMatches, "environment is already bootstrapped")
}

type mockBootstrapInstance struct {
	instance.Instance
}

func (*mockBootstrapInstance) Addresses() ([]network.Address, error) {
	return []network.Address{network.Address{Value: "localhost"}}, nil
}

func (s *BootstrapSuite) TestSeriesDeprecation(c *gc.C) {
	ctx := s.checkSeriesArg(c, "--series")
	c.Check(coretesting.Stderr(ctx), gc.Equals,
		"Use of --series is obsolete. --upload-tools now expands to all supported series of the same operating system.\n")
}

func (s *BootstrapSuite) TestUploadSeriesDeprecation(c *gc.C) {
	ctx := s.checkSeriesArg(c, "--upload-series")
	c.Check(coretesting.Stderr(ctx), gc.Equals,
		"Use of --upload-series is obsolete. --upload-tools now expands to all supported series of the same operating system.\n")
}

func (s *BootstrapSuite) checkSeriesArg(c *gc.C, argVariant string) *cmd.Context {
	_bootstrap := &fakeBootstrapFuncs{}
	s.PatchValue(&getBootstrapFuncs, func() BootstrapInterface {
		return _bootstrap
	})
	resetJujuHome(c)
	s.PatchValue(&allInstances, func(environ environs.Environ) ([]instance.Instance, error) {
		return []instance.Instance{&mockBootstrapInstance{}}, nil
	})

	ctx, err := coretesting.RunCommand(c, envcmd.Wrap(&BootstrapCommand{}), "--upload-tools", argVariant, "foo,bar")

	c.Assert(err, gc.IsNil)
	return ctx
}

// In the case where we cannot examine an environment, we want the
// error to propagate back up to the user.
func (s *BootstrapSuite) TestBootstrapPropagatesEnvErrors(c *gc.C) {

	env := resetJujuHome(c)
	defaultSeriesVersion := version.Current
	defaultSeriesVersion.Series = config.PreferredSeries(env.Config())
	// Force a dev version by having a non zero build number.
	// This is because we have not uploaded any tools and auto
	// upload is only enabled for dev versions.
	defaultSeriesVersion.Build = 1234
	s.PatchValue(&version.Current, defaultSeriesVersion)

	const envName = "peckham"
	_, err := coretesting.RunCommand(c, envcmd.Wrap(&BootstrapCommand{}), "-e", envName)
	c.Assert(err, gc.IsNil)

	// Change permissions on the jenv file to simulate some kind of
	// unexpected error when trying to read info from the environment
	jenvFile := gitjujutesting.HomePath(".juju", "environments", envName+".jenv")
	err = os.Chmod(jenvFile, os.FileMode(0200))
	c.Assert(err, gc.IsNil)

	// The second bootstrap should fail b/c of the propogated error
	_, err = coretesting.RunCommand(c, envcmd.Wrap(&BootstrapCommand{}), "-e", envName)
	c.Assert(err, gc.ErrorMatches, "there was an issue examining the environment: .*")
}

func (s *BootstrapSuite) TestBootstrapJenvWarning(c *gc.C) {
	env := resetJujuHome(c)
	defaultSeriesVersion := version.Current
	defaultSeriesVersion.Series = config.PreferredSeries(env.Config())
	// Force a dev version by having a non zero build number.
	// This is because we have not uploaded any tools and auto
	// upload is only enabled for dev versions.
	defaultSeriesVersion.Build = 1234
	s.PatchValue(&version.Current, defaultSeriesVersion)

	store, err := configstore.Default()
	c.Assert(err, gc.IsNil)
	ctx := coretesting.Context(c)
	environs.PrepareFromName("peckham", ctx, store)

	logger := "jenv.warning.test"
	var testWriter loggo.TestWriter
	loggo.RegisterWriter(logger, &testWriter, loggo.WARNING)
	defer loggo.RemoveWriter(logger)

	_, errc := runCommand(ctx, envcmd.Wrap(new(BootstrapCommand)), "-e", "peckham")
	c.Assert(<-errc, gc.IsNil)
	c.Assert(testWriter.Log(), jc.LogMatches, []string{"ignoring environments.yaml: using bootstrap config in .*"})
}

func (s *BootstrapSuite) TestInvalidLocalSource(c *gc.C) {
	s.PatchValue(&version.Current.Number, version.MustParse("1.2.0"))
	env := resetJujuHome(c)

	// Bootstrap the environment with an invalid source.
	// The command returns with an error.
	_, err := coretesting.RunCommand(c, envcmd.Wrap(&BootstrapCommand{}), "--metadata-source", c.MkDir())
	c.Check(err, gc.ErrorMatches, `failed to bootstrap environment: Juju cannot bootstrap because no tools are available for your environment(.|\n)*`)

	// Now check that there are no tools available.
	_, err = envtools.FindTools(
		env, version.Current.Major, version.Current.Minor, coretools.Filter{}, envtools.DoNotAllowRetry)
	c.Assert(err, gc.FitsTypeOf, errors.NotFoundf(""))
}

// createImageMetadata creates some image metadata in a local directory.
func createImageMetadata(c *gc.C) (string, []*imagemetadata.ImageMetadata) {
	// Generate some image metadata.
	im := []*imagemetadata.ImageMetadata{
		{
			Id:         "1234",
			Arch:       "amd64",
			Version:    "13.04",
			RegionName: "region",
			Endpoint:   "endpoint",
		},
	}
	cloudSpec := &simplestreams.CloudSpec{
		Region:   "region",
		Endpoint: "endpoint",
	}
	sourceDir := c.MkDir()
	sourceStor, err := filestorage.NewFileStorageWriter(sourceDir)
	c.Assert(err, gc.IsNil)
	err = imagemetadata.MergeAndWriteMetadata("raring", im, cloudSpec, sourceStor)
	c.Assert(err, gc.IsNil)
	return sourceDir, im
}

// checkImageMetadata checks that the environment contains the expected image metadata.
func checkImageMetadata(c *gc.C, stor storage.StorageReader, expected []*imagemetadata.ImageMetadata) {
	metadata := imtesting.ParseMetadataFromStorage(c, stor)
	c.Assert(metadata, gc.HasLen, 1)
	c.Assert(expected[0], gc.DeepEquals, metadata[0])
}

func (s *BootstrapSuite) TestUploadLocalImageMetadata(c *gc.C) {
	sourceDir, expected := createImageMetadata(c)
	env := resetJujuHome(c)

	// Bootstrap the environment with the valid source.
	// Force a dev version by having a non zero build number.
	// This is because we have not uploaded any tools and auto
	// upload is only enabled for dev versions.
	devVersion := version.Current
	devVersion.Build = 1234
	s.PatchValue(&version.Current, devVersion)

	_, err := coretesting.RunCommand(c, envcmd.Wrap(&BootstrapCommand{}), "--metadata-source", sourceDir)
	c.Assert(err, gc.IsNil)
	c.Assert(imagemetadata.DefaultBaseURL, gc.Equals, imagemetadata.UbuntuCloudImagesURL)

	// Now check the image metadata has been uploaded.
	checkImageMetadata(c, env.Storage(), expected)
}

func (s *BootstrapSuite) TestValidateConstraintsCalledWithMetadatasource(c *gc.C) {
	sourceDir, _ := createImageMetadata(c)
	resetJujuHome(c)

	// Bootstrap the environment with the valid source.
	// Force a dev version by having a non zero build number.
	// This is because we have not uploaded any tools and auto
	// upload is only enabled for dev versions.
	devVersion := version.Current
	devVersion.Build = 1234
	s.PatchValue(&version.Current, devVersion)

	var calledFuncs []string
	s.PatchValue(&uploadCustomMetadata, func(metadataDir string, env environs.Environ) error {
		c.Assert(metadataDir, gc.DeepEquals, sourceDir)
		calledFuncs = append(calledFuncs, "uploadCustomMetadata")
		return nil
	})
	s.PatchValue(&validateConstraints, func(cons constraints.Value, env environs.Environ) error {
		c.Assert(cons, gc.DeepEquals, constraints.MustParse("mem=4G"))
		calledFuncs = append(calledFuncs, "validateConstraints")
		return nil
	})
	_, err := coretesting.RunCommand(
		c, envcmd.Wrap(&BootstrapCommand{}), "--metadata-source", sourceDir, "--constraints", "mem=4G")
	c.Assert(err, gc.IsNil)
	c.Assert(calledFuncs, gc.DeepEquals, []string{"uploadCustomMetadata", "validateConstraints"})
}

func (s *BootstrapSuite) TestValidateConstraintsCalledWithoutMetadatasource(c *gc.C) {
	// Bootstrap the environment with the valid source.
	// Force a dev version by having a non zero build number.
	// This is because we have not uploaded any tools and auto
	// upload is only enabled for dev versions.
	devVersion := version.Current
	devVersion.Build = 1234
	s.PatchValue(&version.Current, devVersion)

	validateCalled := 0
	s.PatchValue(&validateConstraints, func(cons constraints.Value, env environs.Environ) error {
		c.Assert(cons, gc.DeepEquals, constraints.MustParse("mem=4G"))
		validateCalled++
		return nil
	})
	resetJujuHome(c)
	_, err := coretesting.RunCommand(
		c, envcmd.Wrap(&BootstrapCommand{}), "--constraints", "mem=4G")
	c.Assert(err, gc.IsNil)
	c.Assert(validateCalled, gc.Equals, 1)
}

func (s *BootstrapSuite) TestAutoSyncLocalSource(c *gc.C) {
	sourceDir := createToolsSource(c, vAll)
	s.PatchValue(&version.Current.Number, version.MustParse("1.2.0"))
	env := resetJujuHome(c)

	// Bootstrap the environment with the valid source.
	// The bootstrapping has to show no error, because the tools
	// are automatically synchronized.
	_, err := coretesting.RunCommand(c, envcmd.Wrap(&BootstrapCommand{}), "--metadata-source", sourceDir)
	c.Assert(err, gc.IsNil)

	// Now check the available tools which are the 1.2.0 envtools.
	checkTools(c, env, v120All)
}

func (s *BootstrapSuite) setupAutoUploadTest(c *gc.C, vers, series string) environs.Environ {
	s.PatchValue(&envtools.BundleTools, toolstesting.GetMockBundleTools(c))
	sourceDir := createToolsSource(c, vAll)
	s.PatchValue(&envtools.DefaultBaseURL, sourceDir)

	// Change the tools location to be the test location and also
	// the version and ensure their later restoring.
	// Set the current version to be something for which there are no tools
	// so we can test that an upload is forced.
	s.PatchValue(&version.Current, version.MustParseBinary(vers+"-"+series+"-"+version.Current.Arch))

	// Create home with dummy provider and remove all
	// of its envtools.
	return resetJujuHome(c)
}

func (s *BootstrapSuite) TestAutoUploadAfterFailedSync(c *gc.C) {
	s.PatchValue(&version.Current.Series, config.LatestLtsSeries())
	s.setupAutoUploadTest(c, "1.7.3", "quantal")
	// Run command and check for that upload has been run for tools matching
	// the current juju version.
	opc, errc := runCommand(nullContext(c), envcmd.Wrap(new(BootstrapCommand)))
	c.Assert(<-errc, gc.IsNil)
	c.Check((<-opc).(dummy.OpPutFile).Env, gc.Equals, "peckham") // verify storage
	c.Check((<-opc).(dummy.OpBootstrap).Env, gc.Equals, "peckham")
	mcfg := (<-opc).(dummy.OpFinalizeBootstrap).MachineConfig
	c.Assert(mcfg, gc.NotNil)
	c.Assert(mcfg.Tools.Version.String(), gc.Equals, "1.7.3.1-raring-"+version.Current.Arch)
}

func (s *BootstrapSuite) TestAutoUploadOnlyForDev(c *gc.C) {
	s.setupAutoUploadTest(c, "1.8.3", "precise")
	_, errc := runCommand(nullContext(c), envcmd.Wrap(new(BootstrapCommand)))
	err := <-errc
	c.Assert(err, gc.ErrorMatches,
		"failed to bootstrap environment: Juju cannot bootstrap because no tools are available for your environment(.|\n)*")
}

func (s *BootstrapSuite) TestMissingToolsError(c *gc.C) {
	s.setupAutoUploadTest(c, "1.8.3", "precise")

	_, err := coretesting.RunCommand(c, envcmd.Wrap(&BootstrapCommand{}))
	c.Assert(err, gc.ErrorMatches,
		"failed to bootstrap environment: Juju cannot bootstrap because no tools are available for your environment(.|\n)*")
}

func buildToolsTarballAlwaysFails(forceVersion *version.Number) (*sync.BuiltTools, error) {
	return nil, fmt.Errorf("an error")
}

func (s *BootstrapSuite) TestMissingToolsUploadFailedError(c *gc.C) {
	s.setupAutoUploadTest(c, "1.7.3", "precise")
	s.PatchValue(&sync.BuildToolsTarball, buildToolsTarballAlwaysFails)

	ctx, err := coretesting.RunCommand(c, envcmd.Wrap(&BootstrapCommand{}))

	c.Check(coretesting.Stderr(ctx), gc.Equals, fmt.Sprintf(`
Bootstrapping environment "peckham"
Starting new instance for initial state server
Building tools to upload (1.7.3.1-raring-%s)
`[1:], version.Current.Arch))
	c.Check(err, gc.ErrorMatches, "failed to bootstrap environment: cannot upload bootstrap tools: an error")
}

func (s *BootstrapSuite) TestBootstrapDestroy(c *gc.C) {
	resetJujuHome(c)
	devVersion := version.Current
	// Force a dev version by having a non zero build number.
	// This is because we have not uploaded any tools and auto
	// upload is only enabled for dev versions.
	devVersion.Build = 1234
	s.PatchValue(&version.Current, devVersion)
	opc, errc := runCommand(nullContext(c), envcmd.Wrap(new(BootstrapCommand)), "-e", "brokenenv")
	err := <-errc
	c.Assert(err, gc.ErrorMatches, "failed to bootstrap environment: dummy.Bootstrap is broken")
	var opDestroy *dummy.OpDestroy
	for opDestroy == nil {
		select {
		case op := <-opc:
			switch op := op.(type) {
			case dummy.OpDestroy:
				opDestroy = &op
			}
		default:
			c.Error("expected call to env.Destroy")
			return
		}
	}
	c.Assert(opDestroy.Error, gc.ErrorMatches, "dummy.Destroy is broken")
}

func (s *BootstrapSuite) TestBootstrapKeepBroken(c *gc.C) {
	resetJujuHome(c)
	devVersion := version.Current
	// Force a dev version by having a non zero build number.
	// This is because we have not uploaded any tools and auto
	// upload is only enabled for dev versions.
	devVersion.Build = 1234
	s.PatchValue(&version.Current, devVersion)
	opc, errc := runCommand(nullContext(c), envcmd.Wrap(new(BootstrapCommand)), "-e", "brokenenv", "--keep-broken")
	err := <-errc
	c.Assert(err, gc.ErrorMatches, "failed to bootstrap environment: dummy.Bootstrap is broken")
	done := false
	for !done {
		select {
		case op, ok := <-opc:
			if !ok {
				done = true
				break
			}
			switch op.(type) {
			case dummy.OpDestroy:
				c.Error("unexpected call to env.Destroy")
				break
			}
		default:
			break
		}
	}
}

// createToolsSource writes the mock tools and metadata into a temporary
// directory and returns it.
func createToolsSource(c *gc.C, versions []version.Binary) string {
	versionStrings := make([]string, len(versions))
	for i, vers := range versions {
		versionStrings[i] = vers.String()
	}
	source := c.MkDir()
	toolstesting.MakeTools(c, source, "releases", versionStrings)
	return source
}

// resetJujuHome restores an new, clean Juju home environment without tools.
func resetJujuHome(c *gc.C) environs.Environ {
	jenvDir := gitjujutesting.HomePath(".juju", "environments")
	err := os.RemoveAll(jenvDir)
	c.Assert(err, gc.IsNil)
	coretesting.WriteEnvironments(c, envConfig)
	dummy.Reset()
	store, err := configstore.Default()
	c.Assert(err, gc.IsNil)
	env, err := environs.PrepareFromName("peckham", nullContext(c), store)
	c.Assert(err, gc.IsNil)
	envtesting.RemoveAllTools(c, env)
	return env
}

// checkTools check if the environment contains the passed envtools.
func checkTools(c *gc.C, env environs.Environ, expected []version.Binary) {
	list, err := envtools.FindTools(
		env, version.Current.Major, version.Current.Minor, coretools.Filter{}, envtools.DoNotAllowRetry)
	c.Check(err, gc.IsNil)
	c.Logf("found: " + list.String())
	urls := list.URLs()
	c.Check(urls, gc.HasLen, len(expected))
}

var (
	v100d64 = version.MustParseBinary("1.0.0-raring-amd64")
	v100p64 = version.MustParseBinary("1.0.0-precise-amd64")
	v100q32 = version.MustParseBinary("1.0.0-quantal-i386")
	v100q64 = version.MustParseBinary("1.0.0-quantal-amd64")
	v120d64 = version.MustParseBinary("1.2.0-raring-amd64")
	v120p64 = version.MustParseBinary("1.2.0-precise-amd64")
	v120q32 = version.MustParseBinary("1.2.0-quantal-i386")
	v120q64 = version.MustParseBinary("1.2.0-quantal-amd64")
	v120t32 = version.MustParseBinary("1.2.0-trusty-i386")
	v120t64 = version.MustParseBinary("1.2.0-trusty-amd64")
	v190p32 = version.MustParseBinary("1.9.0-precise-i386")
	v190q64 = version.MustParseBinary("1.9.0-quantal-amd64")
	v200p64 = version.MustParseBinary("2.0.0-precise-amd64")
	v100All = []version.Binary{
		v100d64, v100p64, v100q64, v100q32,
	}
	v120All = []version.Binary{
		v120d64, v120p64, v120q64, v120q32, v120t32, v120t64,
	}
	v190All = []version.Binary{
		v190p32, v190q64,
	}
	v200All = []version.Binary{
		v200p64,
	}
	vAll = joinBinaryVersions(v100All, v120All, v190All, v200All)
)

func joinBinaryVersions(versions ...[]version.Binary) []version.Binary {
	var all []version.Binary
	for _, versions := range versions {
		all = append(all, versions...)
	}
	return all
}

// TODO(menn0): This fake BootstrapInterface implementation is
// currently quite minimal but could be easily extended to cover more
// test scenarios. This could help improve some of the tests in this
// file which execute large amounts of external functionality.
type fakeBootstrapFuncs struct {
}

func (fake *fakeBootstrapFuncs) EnsureNotBootstrapped(env environs.Environ) error {
	return nil
}

func (fake fakeBootstrapFuncs) Bootstrap(ctx environs.BootstrapContext, env environs.Environ, args bootstrap.BootstrapParams) error {
	return nil
}
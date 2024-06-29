package common

import (
	. "github.com/orz-dsh/dsh/utils"
	"testing"
)

func TestParseProjectLinkRegistry(t *testing.T) {
	link, err := ParseProjectLink("registry:")
	if err != nil {
		t.Log(err)
	} else {
		Impossible()
	}

	link, err = ParseProjectLink("registry:foo/bar#ref=tag/v1.0.0")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry fully (1)",
			KV("link", link),
		))
	}
	link, err = ParseProjectLink("registry:foo/bar/1/2/3/4/#ref=tag/v1.0.0")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry fully (2)",
			KV("link", link),
		))
	}

	link, err = ParseProjectLink("registry:foo/bar")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry without ref (1)",
			KV("link", link),
		))
	}
	link, err = ParseProjectLink("registry:foo/bar/1/2/3/4/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry without ref (2)",
			KV("link", link),
		))
	}

	link, err = ParseProjectLink("registry:foo#ref=master")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry without path (1)",
			KV("link", link),
		))
	}
	link, err = ParseProjectLink("registry:foo/#ref=master")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry without path (2)",
			KV("link", link),
		))
	}

	link, err = ParseProjectLink("registry:foo")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry without path and ref (1)",
			KV("link", link),
		))
	}
	link, err = ParseProjectLink("registry:foo/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry without path and ref (2)",
			KV("link", link),
		))
	}
}

func TestParseProjectLinkRegistryAbbr(t *testing.T) {
	link, err := ParseProjectLink("@")
	if err != nil {
		t.Log(err)
	} else {
		Impossible()
	}

	link, err = ParseProjectLink("@foo/bar#ref=tag/v1.0.0")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry abbr fully (1)",
			KV("link", link),
		))
	}
	link, err = ParseProjectLink("@foo/bar/1/2/3/4/#ref=tag/v1.0.0")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry abbr fully (2)",
			KV("link", link),
		))
	}

	link, err = ParseProjectLink("@foo/bar")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry abbr without ref (1)",
			KV("link", link),
		))
	}
	link, err = ParseProjectLink("@foo/bar/1/2/3/4/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry abbr without ref (2)",
			KV("link", link),
		))
	}

	link, err = ParseProjectLink("@foo#ref=master")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry abbr without path (1)",
			KV("link", link),
		))
	}
	link, err = ParseProjectLink("@foo/#ref=master")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry abbr without path (2)",
			KV("link", link),
		))
	}

	link, err = ParseProjectLink("@foo")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry abbr without path and ref (1)",
			KV("link", link),
		))
	}
	link, err = ParseProjectLink("@foo/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link registry abbr without path and ref (2)",
			KV("link", link),
		))
	}
}

func TestParseProjectLinkDir(t *testing.T) {
	link, err := ParseProjectLink("dir:")
	if err != nil {
		t.Log(err)
	} else {
		Impossible()
	}

	link, err = ParseProjectLink("dir:/foo/bar")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link dir (1)",
			KV("link", link),
		))
	}
	link, err = ParseProjectLink("dir:/foo/bar/../../1/2/3/4/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link dir (2)",
			KV("link", link),
		))
	}
}

func TestParseProjectLinkGit(t *testing.T) {
	link, err := ParseProjectLink("git:")
	if err != nil {
		t.Log("link git error", err)
	} else {
		Impossible()
	}

	link, err = ParseProjectLink("git:https://github.com/group/project.git#ref=tag/")
	if err != nil {
		t.Log("link git ref error", err)
	} else {
		Impossible()
	}

	link, err = ParseProjectLink("git:https://github.com/group/project.git#ref=tag/v1.0.0")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link git fully",
			KV("link", link),
		))
	}

	link, err = ParseProjectLink("git:https://github.com/group/project.git")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(DescN("parse link git without ref",
			KV("link", link),
		))
	}
}

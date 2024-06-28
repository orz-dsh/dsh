package core

import "testing"

func TestParseProjectLinkRegistry(t *testing.T) {
	link, err := parseProjectLink("registry:")
	if err != nil {
		t.Log(err)
	} else {
		impossible()
	}

	link, err = parseProjectLink("registry:foo/bar#ref=tag/v1.0.0")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry fully (1)",
			kv("link", link),
		))
	}
	link, err = parseProjectLink("registry:foo/bar/1/2/3/4/#ref=tag/v1.0.0")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry fully (2)",
			kv("link", link),
		))
	}

	link, err = parseProjectLink("registry:foo/bar")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry without ref (1)",
			kv("link", link),
		))
	}
	link, err = parseProjectLink("registry:foo/bar/1/2/3/4/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry without ref (2)",
			kv("link", link),
		))
	}

	link, err = parseProjectLink("registry:foo#ref=master")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry without path (1)",
			kv("link", link),
		))
	}
	link, err = parseProjectLink("registry:foo/#ref=master")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry without path (2)",
			kv("link", link),
		))
	}

	link, err = parseProjectLink("registry:foo")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry without path and ref (1)",
			kv("link", link),
		))
	}
	link, err = parseProjectLink("registry:foo/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry without path and ref (2)",
			kv("link", link),
		))
	}
}

func TestParseProjectLinkRegistryAbbr(t *testing.T) {
	link, err := parseProjectLink("@")
	if err != nil {
		t.Log(err)
	} else {
		impossible()
	}

	link, err = parseProjectLink("@foo/bar#ref=tag/v1.0.0")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry abbr fully (1)",
			kv("link", link),
		))
	}
	link, err = parseProjectLink("@foo/bar/1/2/3/4/#ref=tag/v1.0.0")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry abbr fully (2)",
			kv("link", link),
		))
	}

	link, err = parseProjectLink("@foo/bar")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry abbr without ref (1)",
			kv("link", link),
		))
	}
	link, err = parseProjectLink("@foo/bar/1/2/3/4/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry abbr without ref (2)",
			kv("link", link),
		))
	}

	link, err = parseProjectLink("@foo#ref=master")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry abbr without path (1)",
			kv("link", link),
		))
	}
	link, err = parseProjectLink("@foo/#ref=master")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry abbr without path (2)",
			kv("link", link),
		))
	}

	link, err = parseProjectLink("@foo")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry abbr without path and ref (1)",
			kv("link", link),
		))
	}
	link, err = parseProjectLink("@foo/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link registry abbr without path and ref (2)",
			kv("link", link),
		))
	}
}

func TestParseProjectLinkDir(t *testing.T) {
	link, err := parseProjectLink("dir:")
	if err != nil {
		t.Log(err)
	} else {
		impossible()
	}

	link, err = parseProjectLink("dir:/foo/bar")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link dir (1)",
			kv("link", link),
		))
	}
	link, err = parseProjectLink("dir:/foo/bar/../../1/2/3/4/")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link dir (2)",
			kv("link", link),
		))
	}
}

func TestParseProjectLinkGit(t *testing.T) {
	link, err := parseProjectLink("git:")
	if err != nil {
		t.Log("link git error", err)
	} else {
		impossible()
	}

	link, err = parseProjectLink("git:https://github.com/group/project.git#ref=tag/")
	if err != nil {
		t.Log("link git ref error", err)
	} else {
		impossible()
	}

	link, err = parseProjectLink("git:https://github.com/group/project.git#ref=tag/v1.0.0")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link git fully",
			kv("link", link),
		))
	}

	link, err = parseProjectLink("git:https://github.com/group/project.git")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(desc("parse link git without ref",
			kv("link", link),
		))
	}
}

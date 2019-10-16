package dashboard

import (
	"fmt"
	"testing"
)

func TestTravisJob(t *testing.T) {
	dataMocks := []string{"testdata/travis/branch_passed.json", "testdata/travis/branch_failed.json"}
	expected := []TravisJob{
		{Name: "repo", URL: "https://travis-ci.org/org/repo", Color: "green"},
		{Name: "repo", URL: "https://travis-ci.org/org/repo", Color: "red"},
	}

	for i := range dataMocks {
		server := mockSuccesfulResponseWithBodyFromFile(dataMocks[i], t)

		fmt.Println(server.URL)
		tc := NewTravisClient(server.URL+"/", "")
		repo := TravisRepository{Organization: "org", Name: "repo", Branch: "master"}
		res, err := tc.Job(repo)
		server.Close()

		if err != nil {
			t.Fatal(err)
		}

		if res != expected[i] {
			t.Fatalf("Wrong TravisJob returned. Expected: %v, got: %v", expected[i], res)
		}
	}
}

func TestTravisJob_ConnectionFailed(t *testing.T) {

	tc := NewTravisClient("http://wrongUrl/", "")
	repo := TravisRepository{Organization: "org", Name: "repo", Branch: "master"}
	exp := TravisJob{Name: "repo", URL: "https://travis-ci.org/org/repo", Color: "grey"}
	res, err := tc.Job(repo)

	if err == nil {
		t.Fatal("Expecting an error to be returned, got nil")
	}

	if res != exp {
		t.Fatalf("Wrong TravisJob returned. Expected: %v, got: %v", exp, res)
	}
}

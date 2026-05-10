package core_test

import (
	"testing"

	"github.com/pafthang/servicebase/core"
	basemodels "github.com/pafthang/servicebase/services/base/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	teammodels "github.com/pafthang/servicebase/services/team/models"

	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	"github.com/pafthang/servicebase/tools/list"
)

func TestBaseCollectionEventTags(t *testing.T) {
	c1 := new(collectionmodels.Collection)

	c2 := new(collectionmodels.Collection)
	c2.Id = "a"

	c3 := new(collectionmodels.Collection)
	c3.Name = "b"

	c4 := new(collectionmodels.Collection)
	c4.Id = "a"
	c4.Name = "b"

	scenarios := []struct {
		collection   *collectionmodels.Collection
		expectedTags []string
	}{
		{c1, []string{}},
		{c2, []string{"a"}},
		{c3, []string{"b"}},
		{c4, []string{"a", "b"}},
	}

	for i, s := range scenarios {
		event := new(core.BaseCollectionEvent)
		event.Collection = s.collection

		tags := event.Tags()

		if len(s.expectedTags) != len(tags) {
			t.Fatalf("[%d] Expected %v tags, got %v", i, s.expectedTags, tags)
		}

		for _, tag := range s.expectedTags {
			if !list.ExistInSlice(tag, tags) {
				t.Fatalf("[%d] Expected %v tags, got %v", i, s.expectedTags, tags)
			}
		}
	}
}

func TestModelEventTags(t *testing.T) {
	m1 := new(teammodels.Team)

	c := new(collectionmodels.Collection)
	c.Id = "a"
	c.Name = "b"
	m2 := recordmodels.NewRecord(c)

	scenarios := []struct {
		model        basemodels.Model
		expectedTags []string
	}{
		{m1, []string{"teams"}},
		{m2, []string{"a", "b"}},
	}

	for i, s := range scenarios {
		event := new(core.ModelEvent)
		event.Model = s.model

		tags := event.Tags()

		if len(s.expectedTags) != len(tags) {
			t.Fatalf("[%d] Expected %v tags, got %v", i, s.expectedTags, tags)
		}

		for _, tag := range s.expectedTags {
			if !list.ExistInSlice(tag, tags) {
				t.Fatalf("[%d] Expected %v tags, got %v", i, s.expectedTags, tags)
			}
		}
	}
}

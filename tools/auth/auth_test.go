package auth_test

import (
	"testing"

	"github.com/pafthang/servicebase/tools/auth"
)

func TestNewProviderByName(t *testing.T) {
	var err error
	var p auth.Provider

	// invalid
	p, err = auth.NewProviderByName("invalid")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if p != nil {
		t.Errorf("Expected provider to be nil, got %v", p)
	}

	// github
	p, err = auth.NewProviderByName(auth.NameGithub)
	if err != nil {
		t.Errorf("Expected nil, got error %v", err)
	}
	if _, ok := p.(*auth.Github); !ok {
		t.Error("Expected to be instance of *auth.Github")
	}

	// vk
	p, err = auth.NewProviderByName(auth.NameVK)
	if err != nil {
		t.Errorf("Expected nil, got error %v", err)
	}
	if _, ok := p.(*auth.VK); !ok {
		t.Error("Expected to be instance of *auth.VK")
	}

	// yandex
	p, err = auth.NewProviderByName(auth.NameYandex)
	if err != nil {
		t.Errorf("Expected nil, got error %v", err)
	}
	if _, ok := p.(*auth.Yandex); !ok {
		t.Error("Expected to be instance of *auth.Yandex")
	}

}

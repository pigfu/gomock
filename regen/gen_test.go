package regen

import (
	"regexp"
	"testing"
)

func commonTest(pattern string, t *testing.T) {
	str, err := Generate(pattern)
	if err != nil {
		t.Error(err)
		return
	}
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		t.Error(err)
		return
	}
	if matched {
		t.Logf("str:%s,match!!!", str)
	} else {
		t.Errorf("str:%s,not match!!!", str)
	}
}

func TestNoNoWordBoundary(t *testing.T) {
	pattern := "^great\\B"
	commonTest(pattern, t)
}

func TestEmail(t *testing.T) {
	pattern := "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	commonTest(pattern, t)
}

func TestHttpUrl(t *testing.T) {
	pattern := "^(https?):\\/\\/[-\\w]+(\\.[-\\w]+)*(\\/([\\w\\-\\._\\+\\!]*))*$"
	commonTest(pattern, t)
}

func TestHGUID(t *testing.T) {
	pattern := "[0-9a-fA-F]{8}[-]?([0-9a-fA-F]{4}[-]?){3}[0-9a-fA-F]{12}"
	commonTest(pattern, t)
}

func TestIDCard(t *testing.T) {
	pattern := "(\\d{6})(\\d{4})(\\d{2})(\\d{2})(\\d{3})([0-9]|X)"
	commonTest(pattern, t)
}

func TestWeChatNumber(t *testing.T) {
	pattern := "^[a-zA-Z][-_a-zA-Z0-9]{5,19}$"
	commonTest(pattern, t)
}

func TestChinese(t *testing.T) {
	pattern := "[\u4E00-\u9FA5]{6,}"
	commonTest(pattern, t)
}

func TestGenerate(t *testing.T) {
	pattern := "[1-9]{3}\\.\\d{1,5}"
	commonTest(pattern, t)
}

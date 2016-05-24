package main

import "testing"

func TestResult(t *testing.T) {

	for i := 0; i < 20; i++ {
		mp, _, err := readAndGet("./f.txt")		
		t.Log("card_id:", mp["card_id"])
		if err != nil {
			t.Error(err)
		}

	}

}

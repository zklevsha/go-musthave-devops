package archive

import "testing"

func TestArchive(t *testing.T) {
	name := "testArchive"
	testString := "this is some string"

	t.Run(name, func(t *testing.T) {
		compressed, err := Compress([]byte(testString))
		if err != nil {
			t.Errorf("Failed to compress test string: %s", err.Error())
		}

		decompressed, err := Decompress(compressed)
		if err != nil {
			t.Errorf("Failed to decompress test string: %s", err.Error())
		}

		if string(decompressed) != testString {
			t.Errorf("Decompressed string does not match the original: "+
				"have: %s, want: %s", string(decompressed), testString)
		}
	})

}

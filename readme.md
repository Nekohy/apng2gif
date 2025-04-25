# apng2gif
A pure golang tool for converting APNG to GIF
Just 118 lines of code.
# Usage
```console
apng2gif -i input.png -o output.gif
```

```golang
package main

import (
	"os"
	"github.com/Nekohy/apng2gif"
)

func main() {
	in, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer func(in *os.File) {
		_ = in.Close()
	}(in)

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	_ = apng2gif.Convert(in, out)
}
```
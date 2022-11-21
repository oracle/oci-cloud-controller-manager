package go_ensurefips

import (
	"crypto/tls"
	// this import activates FIPS approved mode, even if there's no
	// BoringCrypto implementation in the executable!
	_ "crypto/tls/fipsonly"
	"debug/elf"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
)

// dummyTLSConfig exists to force the Go compiler into including the
// at least some meaningful BoringCrypto symbols - at the time of this
// writing (check git for the earliest commit that contains this!),
// they're all SHA-related functions and objects - on the assumption
// that the Go compiler either contains all relevant symbols or none
var dummyTLSConfig tls.Config

// boringCryptoSymbolToken matches the pattern BoringCrypto symbols
// should match as described in
// https://go.googlesource.com/go/+/refs/heads/dev.boringcrypto.go1.12/misc/boring/
// and
// https://github.com/golang/go/blob/d003f0850a7d22a2047c1cd6830fca07944f18d1/src/crypto/internal/boring/goboringcrypto.h#L5-L9
const BoringCryptoSymbolToken = "_Cfunc__goboringcrypto_"

// countBoringCryptoSymbols counts the BoringCrypto symbols in the
// provided ELF file; equivalent to:
//
// go tool nm <executable> | grep _Cfunc__goboringcrypto_ | wc -l
func countBoringCryptoSymbols(f *elf.File) (count int, err error) {
	symbols, err := f.Symbols()
	if err != nil {
		err = fmt.Errorf("could not read symbol table: %w", err)
		return
	}

	for _, symbol := range symbols {
		if strings.Contains(symbol.Name, BoringCryptoSymbolToken) {
			count++
		}
	}

	return
}

// WriteFIPSMessage describes a function that can be passed to
// CheckCompliance to output a FIPS success message
type WriteFIPSMessage func(format string, v ...interface{})

// Compliant opens the current executable and then performs checks in
// CheckCompliance; any failure is logged to standard out with an
// "go_ensurefips: " prefix and then forces the process to exit with -1.
//
// Use this in init() inside a main_linux.go to get easy FIPS
// compliance!  If this fails, or you need more flexibility, use
// CheckCompliance, though note you will be responsible for
// documenting and presenting to auditors the format of your FIPS
// compliance attestation.
func Compliant() {
	logger := log.New(os.Stdout, "go_ensurefips: ", log.Ldate|log.Ltime|log.Llongfile)

	// Use /proc/self/exe in case somebody mucked around with os.Args.
	// We do this instead of using dlopen(3) + dlinfo(3) so that we
	// can cope with static executables (you can't dlopen(3) something
	// that isn't dynamically linked, even with a path of NULL).  We
	// also don't attempt to create a pointer into our own address
	// space to avoid difficulties around using debug/elf.
	//
	// Using /proc/self/exe does mean that an executable that deletes
	// itself before Compliant() runs can't be verified, and it may
	// run afoul of other ways to package and run Go programs
	// (e.g. with a ZIP archive at the front.)  At the time of this
	// writing (check git!), we believe most OCI teams don't do
	// anything like this, so we save on complexity at the risk of
	// failures.  Teams that do something truly weird can call
	// CheckCompliance themselves.
	//
	// Finally, we resolve the symlink in Go because os.Open will
	// decide that it's not seekable.  Oops!  This isn't an atomic
	// operation in any event so we lose nothing except the time to do
	// an additional syscall, which should only happen once at startup.
	landmark := "/proc/self/exe"
	executable, err := os.Readlink(landmark)
	if err != nil {
		logger.Fatalf("could not find target of %q: %q", landmark, err)
	}

	f, err := os.Open(executable)
	if err != nil {
		logger.Fatalf("could not open %q for FIPS compliance check: %+v", executable, err)
	}
	defer func() { _ = f.Close() }()

	err = CheckCompliance(executable, f, logger.Printf)
	if err != nil {
		logger.Fatalf("FIPS compliance check failed: %+v", err)
	}
}

// CheckCompliance parses the ELF executable represented by reader and
// located at path -- which should be the source of the running
// executable image -- and performs the following checks:
//
// 1. The executable must be running on an acceptable architecture
// (that is, one we know has a FIPS compliant Go compiler);
//
// 2. It has more than 1 BoringCrypto symbol (see
// BoringCryptoSymbolToken).
//
// On any failure, including the checks above, an informative error is
// returned.  On success, the write function is called to emit a
// success message.
func CheckCompliance(path string, reader io.ReaderAt, write WriteFIPSMessage) error {
	switch runtime.GOARCH {
	case "amd64":
	case "arm64":
	default:
		return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	f, err := elf.NewFile(reader)
	if err != nil {
		return fmt.Errorf("could not parse %q as ELF executable: %w", path, err)
	}

	count, err := countBoringCryptoSymbols(f)
	if err != nil {
		return fmt.Errorf("could not count BoringCrypto symbols in %q: %w", err)
	}

	if count < 1 {
		return fmt.Errorf("too few BoringCrypto symbols found")
	}

	write("FIPS compliance check successful: found %d BoringCrypto symbols", count)

	return nil
}

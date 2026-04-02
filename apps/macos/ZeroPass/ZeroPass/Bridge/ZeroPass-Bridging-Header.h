// Go bridge C header.
// Prefer the generated header from the Xcode build phase (TARGET_TEMP_DIR),
// but fall back to the repo header for indexing if the build output isn't present yet.
#if __has_include("libzeropass.h")
#import "libzeropass.h"
#else
#import "../../../../../bridge/libzeropass.h"
#endif

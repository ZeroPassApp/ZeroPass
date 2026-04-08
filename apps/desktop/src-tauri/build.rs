use std::env;
use std::path::PathBuf;

fn main() {
    // Determine architecture for linking the correct Go C-archive
    let arch = env::var("CARGO_CFG_TARGET_ARCH").unwrap_or_default();
    let lib_suffix = match arch.as_str() {
        "aarch64" => "arm64",
        "x86_64" => "amd64",
        _ => panic!("Unsupported architecture: {arch}"),
    };

    // Path to the bridge directory containing libzeropass_<arch>.a
    let bridge_dir = PathBuf::from(env::var("CARGO_MANIFEST_DIR").unwrap())
        .join("../../../bridge");
    let bridge_dir = bridge_dir
        .canonicalize()
        .unwrap_or_else(|e| panic!("Cannot find bridge dir: {e}"));

    // Link the Go C-archive static library
    println!("cargo:rustc-link-search=native={}", bridge_dir.display());
    println!("cargo:rustc-link-lib=static=zeropass_{lib_suffix}");

    // macOS system frameworks required by Go runtime and the bridge
    println!("cargo:rustc-link-lib=framework=CoreFoundation");
    println!("cargo:rustc-link-lib=framework=Security");
    println!("cargo:rustc-link-lib=resolv");

    // Rebuild if the library changes
    println!(
        "cargo:rerun-if-changed={}/libzeropass_{lib_suffix}.a",
        bridge_dir.display()
    );

    tauri_build::build();
}

import SwiftUI

struct AboutView: View {
    private var versionString: String {
        let v = Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? ""
        let b = Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as? String ?? ""
        if b.isEmpty { return v }
        if v.isEmpty { return b }
        return "\(v) (\(b))"
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("ZeroPass")
                .font(.largeTitle)
                .bold()

            Text("Version \(versionString)")
                .foregroundStyle(.secondary)

            Divider()

            Link("GitHub Repository", destination: URL(string: "https://github.com/zeropass/zeropass")!)
            Text("A local-first credential manager powered by a Go core and a native SwiftUI macOS app.")
                .foregroundStyle(.secondary)

            Spacer()
        }
        .padding(20)
    }
}

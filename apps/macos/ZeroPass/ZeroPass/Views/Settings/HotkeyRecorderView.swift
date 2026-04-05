import AppKit
import Carbon.HIToolbox
import SwiftUI

struct HotkeyRecorderView: View {
    @Environment(\.dismiss) private var dismiss

    @Binding var keyCode: Int
    @Binding var modifiers: Int

    @State private var lastCaptured: String = "Press a key combination…"

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Record Quick Search Hotkey")
                .font(.title3)
                .bold()
                .foregroundStyle(ZPTheme.textPrimary)

            Text(lastCaptured)
                .font(.system(.body, design: .monospaced))
                .foregroundStyle(ZPTheme.textSecondary)

            HotkeyCaptureView { kc, mods in
                keyCode = kc
                modifiers = mods
                lastCaptured = Self.displayString(keyCode: kc, modifiers: mods)
                dismiss()
            }
            .frame(height: 60)
            .background(.regularMaterial)
            .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .stroke(ZPTheme.panelBorderStrong, lineWidth: 1)
            )

            HStack {
                Spacer()
                Button("Cancel") { dismiss() }
            }
        }
        .padding(16)
        .background(ZPTheme.workspaceBackground)
    }

    static func displayString(keyCode: Int, modifiers: Int) -> String {
        let mods = UInt32(modifiers)
        var s = ""
        if (mods & UInt32(cmdKey)) != 0 { s += "⌘" }
        if (mods & UInt32(shiftKey)) != 0 { s += "⇧" }
        if (mods & UInt32(optionKey)) != 0 { s += "⌥" }
        if (mods & UInt32(controlKey)) != 0 { s += "⌃" }
        s += keyName(for: UInt32(keyCode))
        return s
    }

    private static func keyName(for code: UInt32) -> String {
        switch Int(code) {
        case kVK_ANSI_A: return "A"
        case kVK_ANSI_B: return "B"
        case kVK_ANSI_C: return "C"
        case kVK_ANSI_D: return "D"
        case kVK_ANSI_E: return "E"
        case kVK_ANSI_F: return "F"
        case kVK_ANSI_G: return "G"
        case kVK_ANSI_H: return "H"
        case kVK_ANSI_I: return "I"
        case kVK_ANSI_J: return "J"
        case kVK_ANSI_K: return "K"
        case kVK_ANSI_L: return "L"
        case kVK_ANSI_M: return "M"
        case kVK_ANSI_N: return "N"
        case kVK_ANSI_O: return "O"
        case kVK_ANSI_P: return "P"
        case kVK_ANSI_Q: return "Q"
        case kVK_ANSI_R: return "R"
        case kVK_ANSI_S: return "S"
        case kVK_ANSI_T: return "T"
        case kVK_ANSI_U: return "U"
        case kVK_ANSI_V: return "V"
        case kVK_ANSI_W: return "W"
        case kVK_ANSI_X: return "X"
        case kVK_ANSI_Y: return "Y"
        case kVK_ANSI_Z: return "Z"
        default: return "(key \(code))"
        }
    }
}

private struct HotkeyCaptureView: NSViewRepresentable {
    let onCapture: (Int, Int) -> Void

    func makeNSView(context: Context) -> CaptureNSView {
        let v = CaptureNSView()
        v.onCapture = onCapture
        DispatchQueue.main.async {
            v.window?.makeFirstResponder(v)
            _ = v.becomeFirstResponder()
        }
        return v
    }

    func updateNSView(_ nsView: CaptureNSView, context: Context) {}
}

private final class CaptureNSView: NSView {
    var onCapture: ((Int, Int) -> Void)?

    override var acceptsFirstResponder: Bool { true }

    override func keyDown(with event: NSEvent) {
        let mods = event.modifierFlags.intersection([.command, .shift, .option, .control])
        guard !mods.isEmpty else { return }

        let carbonMods = Self.carbonModifiers(from: mods)
        onCapture?(Int(event.keyCode), Int(carbonMods))
    }

    private static func carbonModifiers(from flags: NSEvent.ModifierFlags) -> UInt32 {
        var m: UInt32 = 0
        if flags.contains(.command) { m |= UInt32(cmdKey) }
        if flags.contains(.shift) { m |= UInt32(shiftKey) }
        if flags.contains(.option) { m |= UInt32(optionKey) }
        if flags.contains(.control) { m |= UInt32(controlKey) }
        return m
    }
}

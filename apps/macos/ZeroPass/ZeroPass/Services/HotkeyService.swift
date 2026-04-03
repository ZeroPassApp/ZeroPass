import AppKit
import Carbon.HIToolbox

@MainActor
final class HotkeyService {
    static let shared = HotkeyService()

    struct Hotkey: Equatable {
        var keyCode: UInt32
        var modifiers: UInt32

        static let quickSearchDefault = Hotkey(keyCode: UInt32(kVK_ANSI_K), modifiers: UInt32(cmdKey))
    }

    var onHotkey: (() -> Void)?

    private var hotKeyRef: EventHotKeyRef?
    private var handlerRef: EventHandlerRef?

    private let signature: OSType = 0x5A504853 // 'ZPHS'
    private let hotKeyID: UInt32 = 1

    private init() {}

    func register(_ hotkey: Hotkey) {
        unregister()

        var id = EventHotKeyID(signature: signature, id: hotKeyID)
        let status = RegisterEventHotKey(
            hotkey.keyCode,
            hotkey.modifiers,
            id,
            GetEventDispatcherTarget(),
            0,
            &hotKeyRef
        )

        guard status == noErr else {
            hotKeyRef = nil
            return
        }

        var eventType = EventTypeSpec(eventClass: OSType(kEventClassKeyboard), eventKind: UInt32(kEventHotKeyPressed))

        let callback: EventHandlerUPP = { _, eventRef, _ in
            var hk = EventHotKeyID()
            let err = GetEventParameter(
                eventRef,
                EventParamName(kEventParamDirectObject),
                EventParamType(typeEventHotKeyID),
                nil,
                MemoryLayout<EventHotKeyID>.size,
                nil,
                &hk
            )
            guard err == noErr else { return noErr }
            guard hk.id == HotkeyService.shared.hotKeyID else { return noErr }
            DispatchQueue.main.async {
                HotkeyService.shared.onHotkey?()
            }
            return noErr
        }

        InstallEventHandler(
            GetEventDispatcherTarget(),
            callback,
            1,
            &eventType,
            nil,
            &handlerRef
        )
    }

    func unregister() {
        if let hotKeyRef {
            UnregisterEventHotKey(hotKeyRef)
        }
        hotKeyRef = nil

        if let handlerRef {
            RemoveEventHandler(handlerRef)
        }
        handlerRef = nil
    }
}

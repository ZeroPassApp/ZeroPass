import SwiftUI

struct ItemEditorView: View, Identifiable {
    var id: String { item.id.isEmpty ? UUID().uuidString : item.id }

    @EnvironmentObject var vault: VaultClient
    @Environment(\.dismiss) private var dismiss

    @State var item: VaultItem
    @State private var isBusy = false

    @State private var newFieldKey: String = ""
    @State private var newFieldValue: String = ""

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text(item.id.isEmpty ? "New Item" : "Edit Item")
                .font(.title2)
                .bold()

            Form {
                Picker("Type", selection: $item.type) {
                    ForEach(VaultItemType.allCases) { t in
                        Text(t.displayName).tag(t)
                    }
                }

                TextField("Name", text: $item.name)
                TextField("Notes", text: $item.notes, axis: .vertical)
                    .lineLimit(3...8)

                Toggle("Favorite", isOn: $item.favorite)

                Section("Fields") {
                    ForEach(item.fields.keys.sorted(), id: \.self) { key in
                        TextField(key, text: Binding(
                            get: { item.fields[key] ?? "" },
                            set: { item.fields[key] = $0 }
                        ))
                    }

                    HStack {
                        TextField("Key", text: $newFieldKey)
                        TextField("Value", text: $newFieldValue)
                        Button("Add") {
                            guard !newFieldKey.isEmpty else { return }
                            item.fields[newFieldKey] = newFieldValue
                            newFieldKey = ""
                            newFieldValue = ""
                        }
                    }
                }
            }

            if let err = vault.lastError {
                Text(err)
                    .foregroundStyle(.red)
                    .textSelection(.enabled)
            }

            HStack {
                Button("Cancel") { dismiss() }
                    .disabled(isBusy)

                Spacer()

                Button("Save") { save() }
                    .disabled(isBusy || item.name.isEmpty)
                    .keyboardShortcut(.defaultAction)
            }
        }
        .padding(20)
        .frame(width: 640, height: 560)
    }

    private func save() {
        isBusy = true
        vault.lastError = nil

        Task {
            defer { isBusy = false }
            do {
                try await vault.saveItem(item)
                dismiss()
            } catch {
                vault.lastError = error.localizedDescription
            }
        }
    }
}

import Foundation

struct RecoveryPhraseSummary: Equatable {
    let words: [String]

    var normalizedText: String {
        words.joined(separator: " ")
    }

    var wordCount: Int {
        words.count
    }

    var isEmpty: Bool {
        words.isEmpty
    }

    var isStandardLength: Bool {
        RecoveryPhraseSupport.standardWordCounts.contains(wordCount)
    }

    var helperText: String {
        if isEmpty {
            return "Paste your 12-, 15-, 18-, 21-, or 24-word recovery phrase. Numbering and line breaks are okay."
        }
        if isStandardLength {
            return "\(wordCount) words detected."
        }
        return "\(wordCount) words detected. Recovery phrases usually contain 12, 15, 18, 21, or 24 words."
    }
}

enum RecoveryPhraseSupport {
    static let standardWordCounts: Set<Int> = [12, 15, 18, 21, 24]

    static func summary(from text: String) -> RecoveryPhraseSummary {
        RecoveryPhraseSummary(words: normalizedWords(from: text))
    }

    static func normalizedText(from text: String) -> String {
        summary(from: text).normalizedText
    }

    static func normalizedWords(from text: String) -> [String] {
        text
            .components(separatedBy: .whitespacesAndNewlines)
            .map(cleanToken)
            .filter { !$0.isEmpty }
    }

    private static func cleanToken(_ token: String) -> String {
        let trimmed = token.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return "" }

        let withoutIndex = trimmed.replacingOccurrences(
            of: #"^\d+[\.\):_\-]*"#,
            with: "",
            options: .regularExpression
        )

        return withoutIndex
            .trimmingCharacters(in: CharacterSet.punctuationCharacters.union(.symbols))
            .lowercased()
    }
}

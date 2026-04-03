import XCTest
@testable import ZeroPass

final class RecoveryPhraseSupportTests: XCTestCase {
    func testSummaryNormalizesLineBreaksAndNumbering() {
        let input = """
        1. Alpha
        2. Beta
        3) Gamma
        4: Delta
        """

        let summary = RecoveryPhraseSupport.summary(from: input)

        XCTAssertEqual(summary.words, ["alpha", "beta", "gamma", "delta"])
        XCTAssertEqual(summary.normalizedText, "alpha beta gamma delta")
    }

    func testSummaryRecognizesStandardPhraseLengths() {
        let input = (1...12)
            .map { "word\($0)" }
            .joined(separator: " ")

        let summary = RecoveryPhraseSupport.summary(from: input)

        XCTAssertEqual(summary.wordCount, 12)
        XCTAssertTrue(summary.isStandardLength)
    }
}

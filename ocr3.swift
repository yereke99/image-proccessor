import Foundation
import Vision

// Define the OCR function
func performOCR(on image: CGImage, usesCPUOnly: Bool) -> String? {

    let request = VNRecognizeTextRequest(completionHandler: { request, error in
        if let error = error {
            print("Error: \(error)")
            return
        }
        guard let observations = request.results as? [VNRecognizedTextObservation] else {
            return
        }
        let recognizedText = observations.compactMap { observation in
            return observation.topCandidates(1).first?.string
        }.joined(separator: "\n")
        print(recognizedText)
    })
    request.recognitionLevel = .accurate
    request.usesLanguageCorrection = true
    request.usesCPUOnly = usesCPUOnly
    let handler = VNImageRequestHandler(cgImage: image, options: [:])
    do {
        try handler.perform([request])
        return request.results?.description
    } catch {
        print("Error: \(error)")
        return nil
    }
}

// Parse the command-line arguments
guard CommandLine.arguments.count == 3 else {
    print("Usage: ocr <image_path> <uses_cpu_only>")
    exit(5)
}
let imagePath = CommandLine.arguments[1]
let usesCPUOnly = Bool(CommandLine.arguments[2]) ?? true

// Load the image and perform OCR
let fileUrl = URL(fileURLWithPath: imagePath)
guard let imageSource = CGImageSourceCreateWithURL(fileUrl as CFURL, nil) else {
    print("Error: failed to create image source for \(imagePath)")
    exit(2)
}
let imageOptions: [CFString: Any] = [
    kCGImageSourceShouldCache: false,
    kCGImageSourceShouldAllowFloat: false
]
guard let image = CGImageSourceCreateImageAtIndex(imageSource, 0, imageOptions as CFDictionary) else {
    print("Error: failed to create image from \(imagePath)")
    exit(3)
}
if let recognizedText = performOCR(on: image, usesCPUOnly: usesCPUOnly) {
    print("Recognized text: \(recognizedText)")
} else {
    exit(4)
}
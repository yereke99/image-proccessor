import Foundation
import Vision
import CoreImage

// Функция для распознавания текста на изображении
func recognizeText(imageURL: URL) {
    guard let image = CIImage(contentsOf: imageURL),
          let cgImage = convertCIImageToCGImage(image) else {
        print("failed to load or edit image")
        return
    }

    let textRecognitionRequest = VNRecognizeTextRequest { request, error in
        guard let observations = request.results as? [VNRecognizedTextObservation] else {
            print("error during text recognition:", error?.localizedDescription ?? "unknown error")
            return
        }

        // Получение распознанного текста
        let recognizedText = observations.compactMap { observation in
            observation.topCandidates(1).first?.string
        }.joined(separator: "\n")

        print(recognizedText)
    }

    let requestHandler = VNImageRequestHandler(cgImage: cgImage)
    do {
        try requestHandler.perform([textRecognitionRequest])
    } catch {
        print("error executing request:", error.localizedDescription)
    }
}

// Преобразование CIImage в CGImage
func convertCIImageToCGImage(_ image: CIImage) -> CGImage? {
    let context = CIContext(options: nil)
    return context.createCGImage(image, from: image.extent)
}

// Получение пути к изображению из аргументов командной строки
guard CommandLine.arguments.count >= 2 else {
    print("specify the path to the image in the command line arguments")
    exit(1)
}

let imagePath = CommandLine.arguments[1]
let imageURL = URL(fileURLWithPath: imagePath)

recognizeText(imageURL: imageURL)

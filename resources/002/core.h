#pragma once

#include "Optimizer.h"
#include "Layer.h"

#include <vector>
#include <map>
#include <cmath>
#include <sstream>
#include <iostream>
#include <cstdlib>
#include <cmath>
#include <string>

// Below are the core optimizers components

struct AdamWeights {
    std::vector<std::vector<double>> m_weights; // First moment for weights (2D)
    std::vector<std::vector<double>> v_weights; // Second moment for weights (2D)
    std::vector<double> m_biases;               // First moment for biases (1D)
    std::vector<double> v_biases;               // Second moment for biases (1D)
};
class Adam : public Optimizer {
private:
    double beta1;
    double beta2;
    double epsilon;
    int t; // Time step
    std::map<std::string, AdamWeights> adamWeights;

public:
    Adam() :Adam(0.0001) {
    };
    Adam(double lr, double beta1 = 0.9, double beta2 = 0.999, double epsilon = 1e-8) : beta1(beta1), beta2(beta2), epsilon(epsilon), t(0) {
        learningRate = lr;
    };

    void initialize(Layer* layer) {
        const auto& weights = layer->getWeights();
        const auto& biases = layer->getBiases();

        adamWeights[layer->name] = AdamWeights();
        if (weights.size() == 0) {
            return;
        }

        adamWeights[layer->name].m_weights.resize(weights.size(), std::vector<double>(weights[0].size(), 0.0));
        adamWeights[layer->name].v_weights.resize(weights.size(), std::vector<double>(weights[0].size(), 0.0));
        adamWeights[layer->name].m_biases.resize(biases.size(), 0.0);
        adamWeights[layer->name].v_biases.resize(biases.size(), 0.0);
    };
    void update(std::vector<Layer*>& layers) override {
        double lr = learningRate;
        t++; // Increment the time step

        for (auto layer : layers) {
            auto& weights = layer->getWeights();
            auto& biases = layer->getBiases();
            auto& gradWeights = layer->getGradWeights();
            auto& gradBiases = layer->getGradBiases();

            if (adamWeights.count(layer->name) == 0) {
                initialize(layer); // Initialize moments if not already done
            }

            // Update weights
            for (size_t i = 0; i < weights.size(); i++) {
                for (size_t j = 0; j < weights[i].size(); j++) {
                    adamWeights[layer->name].m_weights[i][j] = beta1 * adamWeights[layer->name].m_weights[i][j] + (1 - beta1) * gradWeights[i][j];
                    adamWeights[layer->name].v_weights[i][j] = beta2 * adamWeights[layer->name].v_weights[i][j] + (1 - beta2) * gradWeights[i][j] * gradWeights[i][j];

                    // Bias correction
                    double m_hat = adamWeights[layer->name].m_weights[i][j] / (1 - pow(beta1, t));
                    double v_hat = adamWeights[layer->name].v_weights[i][j] / (1 - pow(beta2, t));

                    // Update the parameter
                    weights[i][j] -= lr * m_hat / (sqrt(v_hat) + epsilon);
                }
            }

            // Update biases
            for (size_t i = 0; i < biases.size(); i++) {
                adamWeights[layer->name].m_biases[i] = beta1 * adamWeights[layer->name].m_biases[i] + (1 - beta1) * gradBiases[i];
                adamWeights[layer->name].v_biases[i] = beta2 * adamWeights[layer->name].v_biases[i] + (1 - beta2) * gradBiases[i] * gradBiases[i];

                // Bias correction
                double m_hat = adamWeights[layer->name].m_biases[i] / (1 - pow(beta1, t));
                double v_hat = adamWeights[layer->name].v_biases[i] / (1 - pow(beta2, t));

                // Update the parameter
                biases[i] -= lr * m_hat / (sqrt(v_hat) + epsilon);
            }
        }
    };

    std::string saveToString() override {
        std::ostringstream oss;

        // Save general Adam hyperparameters
        oss << "Adam " << learningRate << " " << beta1 << " " << beta2 << " " << epsilon << " " << t << "\n";

        // Save all moment vectors stored in adamWeights
        for (const auto& pair : adamWeights) {
            const std::string& layerName = pair.first;
            const AdamWeights& weights = pair.second;

            oss << layerName << "\n";
            if (weights.m_weights.size() == 0) {
                oss << 0 << " " << 0 << " " << 0 << "\n";
                continue;
            }
            else {
                oss << weights.m_weights.size() << " " << weights.m_weights[0].size() << " " << weights.m_biases.size() << "\n";
            }

            // Save m_weights
            for (const auto& row : weights.m_weights) {
                for (double val : row) {
                    oss << val << " ";
                }
                oss << "\n";
            }

            // Save v_weights
            for (const auto& row : weights.v_weights) {
                for (double val : row) {
                    oss << val << " ";
                }
                oss << "\n";
            }

            // Save m_biases
            for (double val : weights.m_biases) {
                oss << val << " ";
            }
            oss << "\n";

            // Save v_biases
            for (double val : weights.v_biases) {
                oss << val << " ";
            }
            oss << "\n";
        }
        oss << "\n";

        return oss.str();
    };
    void loadFromString(const std::string& str) override {
        std::istringstream iss(str);

        // Load general Adam hyperparameters
        std::string optimizerType;
        iss >> optimizerType;

        if (optimizerType != "Adam") {
            throw std::runtime_error("Invalid optimizer type: " + optimizerType);
        }

        iss >> learningRate >> beta1 >> beta2 >> epsilon >> t;

        // Load moment vectors for each layer
        std::string layerName;
        while (iss >> layerName) {
            AdamWeights weights;

            size_t numWeightsRows, numWeightsCols, numBiases;
            iss >> numWeightsRows >> numWeightsCols >> numBiases;

            // Load m_weights
            weights.m_weights.resize(numWeightsRows, std::vector<double>(numWeightsCols, 0.0));
            for (auto& row : weights.m_weights) {
                for (double& val : row) {
                    iss >> val;
                }
            }

            // Load v_weights
            weights.v_weights.resize(numWeightsRows, std::vector<double>(numWeightsCols, 0.0));
            for (auto& row : weights.v_weights) {
                for (double& val : row) {
                    iss >> val;
                }
            }

            // Load m_biases
            weights.m_biases.resize(numBiases, 0.0);
            for (double& val : weights.m_biases) {
                iss >> val;
            }

            // Load v_biases
            weights.v_biases.resize(numBiases, 0.0);
            for (double& val : weights.v_biases) {
                iss >> val;
            }

            // Save weights for the corresponding layer
            adamWeights[layerName] = weights;
        }
    };

    void print() override {
        std::cout << "(Optimizer Type: Adam)" << std::endl;
    };
};

class SGD : public Optimizer {
public:
    SGD() :SGD(0.001) {};
    SGD(double lr) {
        learningRate = lr;
    };
    void update(std::vector<Layer*>& layers) override {
        double lr = learningRate; // Default SGD learning rate

        for (auto layer : layers) {

            auto& weights = layer->getWeights();
            auto& biases = layer->getBiases();
            auto& gradWeights = layer->getGradWeights();
            auto& gradBiases = layer->getGradBiases();

            // Update weights
            for (size_t i = 0; i < weights.size(); i++) {
                for (size_t j = 0; j < weights[i].size(); j++) {
                    weights[i][j] -= lr * gradWeights[i][j];
                }
            }

            // Update biases
            for (size_t i = 0; i < biases.size(); i++) {
                biases[i] -= lr * gradBiases[i];
            }
        }
    };

    std::string saveToString() override {
        std::ostringstream oss;
        oss << "SGD " << learningRate << "\n";
        return oss.str();
    };
    void loadFromString(const std::string& str) override {
        std::istringstream iss(str);
        std::string type;
        iss >> type;

        if (type != "SGD") {
            throw std::runtime_error("Invalid optimizer type: " + type);
        }

        iss >> learningRate;
    };

    void print() override {
        std::cout << "(Optimizer Type: SGD)" << std::endl;
    };
};


// Below are the core Layers components

class Dense : public Layer {
private:
    int inputSize;
    int outputSize;
    std::vector<std::vector<double>> weights;
    std::vector<double> biases;
    std::vector<std::vector<double>> gradWeights;
    std::vector<double> gradBiases;
    std::vector<double> lastInput;

public:
    Dense(int inputSize, int outputSize, std::string layerName) : inputSize(inputSize), outputSize(outputSize) {
        name = layerName;
        weights.resize(outputSize, std::vector<double>(inputSize));
        biases.resize(outputSize, 0.0);
        gradWeights.resize(outputSize, std::vector<double>(inputSize, 0.0));
        gradBiases.resize(outputSize, 0.0);

        // Initialize weights and biases with small random values
        for (int j = 0; j < outputSize; j++) {
            for (int i = 0; i < inputSize; i++) {
                weights[j][i] = ((double)rand() / RAND_MAX) * 0.2 - 0.1; // Random values between -0.1 and 0.1
            }
            biases[j] = ((double)rand() / RAND_MAX) * 0.2 - 0.1;
        }
    };
    // Forward pass: Applies the layer to the input
    std::vector<double> forward(const std::vector<double>& input, bool train = true) override {
        lastInput = input;
        std::vector<double> output(outputSize, 0.0);

        for (int j = 0; j < outputSize; j++) {
            double sum = 0.0;
            for (int i = 0; i < inputSize; i++) {
                sum += weights[j][i] * input[i];
            }
            output[j] = sum + biases[j];
        }

        return output;
    };
    // Backward pass: Computes gradients for backpropagation
    std::vector<double> backward(const std::vector<double>& gradOutput) override {
        std::vector<double> gradInput(inputSize, 0.0);

        // Gradients for weights and biases
        for (int j = 0; j < outputSize; j++) {
            for (int i = 0; i < inputSize; i++) {
                gradWeights[j][i] = gradOutput[j] * lastInput[i];
            }
            gradBiases[j] = gradOutput[j];
        }

        // Gradients for input
        for (int i = 0; i < inputSize; i++) {
            for (int j = 0; j < outputSize; j++) {
                gradInput[i] += gradOutput[j] * weights[j][i];
            }
        }

        return gradInput;
    };

    std::vector<std::vector<double>>& getWeights() override { return weights; }
    std::vector<double>& getBiases() override { return biases; }
    std::vector<std::vector<double>>& getGradWeights() override { return gradWeights; }
    std::vector<double>& getGradBiases() override { return gradBiases; }

    // Save the layer's parameters as a string
    std::string saveToString() override {
        std::ostringstream oss;
        oss << "Dense " << inputSize << " " << outputSize << " " << name << "\n";

        for (const auto& row : weights) {
            for (double w : row) {
                oss << w << " ";
            }
            oss << "\n";
        }

        for (double b : biases) {
            oss << b << " ";
        }
        oss << "\n";
        return oss.str();
    };
    // Load the layer's parameters from a string
    void loadFromString(const std::string& str) override {
        std::istringstream iss(str);
        std::string type;
        iss >> type;

        if (type != "Dense") {
            throw std::runtime_error("Invalid layer type: " + type);
        }

        iss >> inputSize >> outputSize >> name;

        weights.resize(outputSize, std::vector<double>(inputSize));
        biases.resize(outputSize);

        for (auto& row : weights) {
            for (double& w : row) {
                iss >> w;
            }
        }

        for (double& b : biases) {
            iss >> b;
        }
    };

    void print() override {
        std::cout << "(Layer Type: Dense, Lyaer Name: " << name << ", Input size: " << inputSize << ", Output size:" << outputSize << ")" << std::endl;
    };
};

class ReLU : public Layer {
private:
    std::vector<double> lastInput;

public:
    ReLU(std::string layerName) { name = layerName; };
    // Forward pass: Applies the layer to the input
    std::vector<double> forward(const std::vector<double>& input, bool train = true) override {
        lastInput = input;
        std::vector<double> output(input.size());

        for (size_t i = 0; i < input.size(); i++) {
            output[i] = std::max(0.0, input[i]);
        }

        return output;
    };
    // Backward pass: Computes gradients for backpropagation
    std::vector<double> backward(const std::vector<double>& gradOutput) override {
        std::vector<double> gradInput(gradOutput.size());

        for (size_t i = 0; i < gradOutput.size(); i++) {
            gradInput[i] = gradOutput[i] * (lastInput[i] > 0 ? 1.0 : 0.0);
        }

        return gradInput;
    };

    std::vector<std::vector<double>>& getWeights() override { return *(new std::vector<std::vector<double>>()); }
    std::vector<double>& getBiases() override { return *(new std::vector<double>()); }
    std::vector<std::vector<double>>& getGradWeights() override { return *(new std::vector<std::vector<double>>()); }
    std::vector<double>& getGradBiases() override { return *(new std::vector<double>()); }

    // Save the state of the Sigmoid layer (though it has no parameters)
    std::string saveToString() override {
        return "ReLU " + name + "\n"; // Sigmoid has no parameters, just return its type
    };
    // Load the state of the Sigmoid layer (though it has no parameters)
    void loadFromString(const std::string& str) override {
        std::istringstream iss(str);
        std::string type;
        iss >> type >> name;

        if (type != "ReLU") {
            throw std::runtime_error("Invalid layer type: " + type);
        }
        // No parameters to load for Sigmoid
    };

    void print() override {
        std::cout << "(Layer Type: ReLU)" << std::endl;
    };
};

class Sigmoid : public Layer {
private:
    std::vector<double> lastOutput; // Stores the output of the forward pass for use in backpropagation

public:
    Sigmoid(std::string layerName) { name = layerName; };
    // Forward pass: Applies the layer to the input
    std::vector<double> forward(const std::vector<double>& input, bool train = true) override {
        lastOutput.resize(input.size());
        for (size_t i = 0; i < input.size(); i++) {
            lastOutput[i] = 1.0 / (1.0 + std::exp(-input[i])); // Sigmoid formula: 1 / (1 + e^(-x))
        }
        return lastOutput;
    };
    // Backward pass: Computes gradients for backpropagation
    std::vector<double> backward(const std::vector<double>& gradOutput) override {
        std::vector<double> gradInput(gradOutput.size());
        for (size_t i = 0; i < gradOutput.size(); i++) {
            double sigmoid_derivative = lastOutput[i] * (1.0 - lastOutput[i]); // Derivative of sigmoid: sigmoid(x) * (1 - sigmoid(x))
            gradInput[i] = gradOutput[i] * sigmoid_derivative;
        }
        return gradInput;
    };

    std::vector<std::vector<double>>& getWeights() override { return *(new std::vector<std::vector<double>>()); }
    std::vector<double>& getBiases() override { return *(new std::vector<double>()); }
    std::vector<std::vector<double>>& getGradWeights() override { return *(new std::vector<std::vector<double>>()); }
    std::vector<double>& getGradBiases() override { return *(new std::vector<double>()); }

    // Save the state of the Sigmoid layer (though it has no parameters)
    std::string saveToString() override {
        return "Sigmoid " + name + "\n"; // Sigmoid has no parameters, just return its type
    };
    // Load the state of the Sigmoid layer (though it has no parameters)
    void loadFromString(const std::string& str) override {
        std::istringstream iss(str);
        std::string type;
        iss >> type >> name;

        if (type != "Sigmoid") {
            throw std::runtime_error("Invalid layer type: " + type);
        }
        // No parameters to load for Sigmoid
    };

    void print() override {
        std::cout << "(Layer Type: Sigmoid)" << std::endl;
    };
};


// Below are some other core components

class Model {
private:
    std::vector<Layer*> layers;
    Optimizer* optimizer;

public:
    void addLayer(Layer* layer) {
        layers.push_back(layer);
    };
    void compile(Optimizer* opt) {
        optimizer = opt;
    };
    std::vector<double> forward(const std::vector<double>& input, bool train = false) {
        std::vector<double> output = input;

        for (auto* layer : layers) {
            output = layer->forward(output, train);
        }

        return output;
    };
    void train(const std::vector<std::vector<double>>& x_train,
        const std::vector<std::vector<double>>& y_train,
        int epochs, std::string saveFileName, bool saveBestOnly = true) {
        double minLoss = -1;
        for (int epoch = 0; epoch < epochs; epoch++) {
            for (size_t i = 0; i < x_train.size(); i++) {
                // Forward pass
                std::vector<double> output = forward(x_train[i], true);

                // Compute loss (MSE)
                double loss = 0.0;
                std::vector<double> gradOutput(output.size());
                for (size_t j = 0; j < output.size(); j++) {
                    double diff = output[j] - y_train[i][j];
                    loss += 0.5 * diff * diff;
                    gradOutput[j] = diff;
                }

                // Backpropagation
                std::vector<double> grad = gradOutput;
                for (auto it = layers.rbegin(); it != layers.rend(); ++it) {
                    grad = (*it)->backward(grad);
                }

                // Update weights
                optimizer->update(layers);
            }
            double loss = valid(x_train, y_train);
            std::cout << "Epoch " << epoch + 1 << "/" << epochs << ", Loss: " << loss << std::endl;
            if (saveBestOnly) {
                if ((minLoss == -1 || loss < minLoss)) {
                    minLoss = loss;
                    save(saveFileName);
                }
            }
            else {
                save(saveFileName);
            }
        }
        load(saveFileName);
    };
    double valid(const std::vector<std::vector<double>>& x_train,
        const std::vector<std::vector<double>>& y_train) {
        double totalLoss = 0.;
        for (size_t i = 0; i < x_train.size(); i++) {
            // Forward pass
            std::vector<double> output = forward(x_train[i], true);

            // Compute loss (MSE)
            double loss = 0.0;
            for (size_t j = 0; j < output.size(); j++) {
                double diff = output[j] - y_train[i][j];
                loss += 0.5 * diff * diff;
            }
            totalLoss += loss;
        }
        return totalLoss / x_train.size();
    };
    std::vector<double> test(const std::vector<double>& x) {
        double minLoss = -1;
        std::vector<double> output = forward(x, true);
        return output;
    };
    void print() {
        for (auto layer : layers) {
            layer->print();
        }
        std::cout << std::endl;
        optimizer->print();
    };

    // Save and load functions
    void save(const std::string& filename, bool silent = false) {
        std::ofstream outFile(filename, std::ios::out);

        if (!outFile.is_open()) {
            std::cout << "Failed to open file for saving: " << filename << std::endl;
            return;
        }
        outFile << "#include <string>\nstd::string model_weights = R\"(\n";

        // Save all layers
        outFile << layers.size() << "\n";
        for (auto* layer : layers) {
            outFile << layer->saveToString() << "\n";
        }

        // Save optimizer
        if (optimizer) {
            outFile << optimizer->saveToString() << "\n";
        }

        outFile << ")\"\n";

        outFile.close();
        if (!silent) std::cout << "\033[1;32mModel saved to " << filename << "\033[0m" << std::endl;
    };
    bool load(const std::string& filename, bool silent = false, bool loadOptimizer = true) {
        try {
            std::ifstream inFile(filename, std::ios::in);

            if (!inFile.is_open()) {
                std::cout << "Failed to open file for loading: " << filename << std::endl;
                return false;
            }

            std::string nullStr;
            std::getline(inFile, nullStr);
            std::getline(inFile, nullStr);

            size_t numLayers;
            inFile >> numLayers;
            inFile.ignore(); // Ignore newline after reading the number of layers

            if (numLayers != layers.size()) {
                throw std::runtime_error("Mismatch between number of layers in file and the model.");
            }

            for (size_t i = 0; i < numLayers; i++) {
                // Read layer data as multiline input
                std::ostringstream layerDataStream;
                std::string line;

                while (std::getline(inFile, line)) {
                    // Empty line indicates end of layer data
                    if (line.empty()) break;

                    layerDataStream << line << "\n";
                }

                // Apply the read data to the corresponding layer
                layers[i]->loadFromString(layerDataStream.str());
            }

            // Load optimizer state
            if (optimizer && loadOptimizer) {
                std::ostringstream optimizerDataStream;
                std::string line;

                while (std::getline(inFile, line)) {
                    // Empty line indicates end of optimizer data
                    if (line.empty()) break;

                    optimizerDataStream << line << "\n";
                }

                optimizer->loadFromString(optimizerDataStream.str());
            }

            inFile.close();
            if (!silent) std::cout << "\033[1;34mModel weights and optimizer state loaded from " << filename << "\033[0m" << std::endl;
            return true;
        }
        catch (...) {
            return false;
        }
    };
};
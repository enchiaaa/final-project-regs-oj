#ifndef LAYER_H
#define LAYER_H

#include <vector>
#include <string>

class Layer {
public:
    std::string name = "";
    // Save the layer state to a string
    virtual std::string saveToString() = 0;
    // Load the layer state from a string
    virtual void loadFromString(const std::string& str) = 0;
    virtual void print() = 0;

    virtual std::vector<std::vector<double>>& getWeights() = 0;
    virtual std::vector<double>& getBiases() = 0;
    virtual std::vector<std::vector<double>>& getGradWeights() = 0;
    virtual std::vector<double>& getGradBiases() = 0;


    virtual std::vector<double> forward(const std::vector<double>& input, bool train = true) = 0;
    virtual std::vector<double> backward(const std::vector<double>& gradOutput) = 0;
};

#endif // LAYER_H

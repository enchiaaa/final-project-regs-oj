#ifndef OPTIMIZER_H
#define OPTIMIZER_H

#include "Layer.h"
#include <map>
#include <string>

class Optimizer {
public:
    double learningRate = 0.001;
    // Save the optimizer state to a string
    virtual std::string saveToString() = 0;
    // Load the optimizer state from a string
    virtual void loadFromString(const std::string& str) = 0;

    virtual void print() = 0;

    virtual void update(std::vector<Layer*>& layers) = 0;
};

#endif // OPTIMIZER_H

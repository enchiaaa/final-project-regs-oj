#include <iostream>
#include <fstream>
#include <vector>
#include <cmath>
#include "core.h"

// DATASET_NAME: {cosine, iris, housing, human_age, cancer_type}
const std::string DATASET_NAME = "cosine";
// Use Adam or SGD optimizer
const std::string OPTIMIZER = "SGD";
// Reference lr of Adam: [0.0001, 0.001]
// Reference lr of SGD: [0.001, 0.01]
#define LEARNING_RATE 0.01
// Reference epochs: [100, 200, 1000]
#define EPOCHS 200

// Requirements of each dataset:
// cosine:      loss < 0.01
// iris:        loss < 0.015
// housing:     loss < 0.04
// human_age:   loss < 0.01
// cancer_type: loss < 0.04

std::pair<std::vector<std::vector<double>>, std::vector<std::vector<double>>> loadData(const std::string& filename) {
	std::vector<std::vector<double>> x_train;
	std::vector<std::vector<double>> y_train;
	std::ifstream inFile(filename, std::ios::in);

	if (!inFile.is_open()) {
		std::cout << "Failed to open file for loading: " << filename << std::endl;
		return std::pair<std::vector<std::vector<double>>, std::vector<std::vector<double>>>();
	}

	std::string line, line2;
	std::getline(inFile, line); // x columns name
	std::getline(inFile, line); // y columns name

	while (std::getline(inFile, line)) {
		std::getline(inFile, line2);
		std::vector<double> x;
		std::vector<double> y;
		std::istringstream iss(line);
		std::istringstream iss2(line2);
		double data;
		while (iss >> data) x.push_back(data);
		while (iss2 >> data) y.push_back(data);
		x_train.push_back(x);
		y_train.push_back(y);
	}
	return std::pair<std::vector<std::vector<double>>, std::vector<std::vector<double>>>(x_train, y_train);
}

Model buildModel(int input_dim, int output_dim, int hidden_dim, bool addSigmoid) {
	Model model;
	model.addLayer(new Dense(input_dim, hidden_dim, "DenseLayer1"));
	model.addLayer(new ReLU("ReluLayer1"));
	model.addLayer(new Dense(hidden_dim, hidden_dim, "DenseLayer2"));
	model.addLayer(new ReLU("ReluLayer2"));
	model.addLayer(new Dense(hidden_dim, output_dim, "DenseLayer2"));
	if (addSigmoid) {
		model.addLayer(new Sigmoid("SigmoidLayer1"));
	}
	return model;
}

Model buildModelCosin(int input_dim, int output_dim) {
	int hidden_dim = 20;
	return buildModel(input_dim, output_dim, hidden_dim, false);
}

Model buildModelIris(int input_dim, int output_dim) {
	int hidden_dim = 20;
	return buildModel(input_dim, output_dim, hidden_dim, false);
}

Model buildModelHousing(int input_dim, int output_dim) {
	int hidden_dim = 20;
	return buildModel(input_dim, output_dim, hidden_dim, false);
}

Model buildModelHumanAge(int input_dim, int output_dim) {
	int hidden_dim = 20;
	return buildModel(input_dim, output_dim, hidden_dim, false);
}

Model buildModelCancerType(int input_dim, int output_dim) {
	int hidden_dim = 20;
	return buildModel(input_dim, output_dim, hidden_dim, true);
}

int main() {
	srand(0); // using seed to make sure the result of your model are stable for OJ to judge

	// Read training data
	std::vector<std::vector<double>> x_train;
	std::vector<std::vector<double>> y_train;
	auto trainData = loadData(DATASET_NAME + ".txt");
	x_train = trainData.first;
	y_train = trainData.second;

	// Set up the neural network model
	int input_dim = (int)x_train[0].size();
	int output_dim = (int)y_train[0].size();

	Model model;
	if (DATASET_NAME == "cosine") model = buildModelCosin(input_dim, output_dim);
	else if (DATASET_NAME == "iris")  model = buildModelIris(input_dim, output_dim);
	else if (DATASET_NAME == "housing")  model = buildModelHousing(input_dim, output_dim);
	else if (DATASET_NAME == "human_age")  model = buildModelHumanAge(input_dim, output_dim);
	else if (DATASET_NAME == "cancer_type")  model = buildModelCancerType(input_dim, output_dim);

	// Use Adam or SGD optimizer
	Optimizer* optimizer = new SGD(LEARNING_RATE); // default using SGD;
	if (OPTIMIZER == "Adam") optimizer = new Adam(LEARNING_RATE);
	else if (OPTIMIZER == "SGD") optimizer = new SGD(LEARNING_RATE);
	else {
		std::cout << "Wrong Optimizer!!" << std::endl;
		return -1;
	}

	model.compile(optimizer);

	// Train the model
	int epochs = EPOCHS; // Increased epochs for better convergence
	model.train(x_train, y_train, epochs, DATASET_NAME + "_model.h");

	std::cout << "\n============================================\n";
	std::cout << "\033[1;36mModel Structure:\033[0m\n";
	model.print();

	// Test the model with a sample input
	std::cout << "\n============================================\n";
	double finalLoss = model.valid(x_train, y_train);
	for (int i = 0; i < x_train.size(); i += 19) {
		std::vector<double> test_input = x_train[i];
		std::vector<double> output = model.forward(test_input);
		std::vector<double> true_data = y_train[i];
		// Print the test result
		std::cout << "Test input: ";
		for (int j = 0; j < test_input.size(); j++) std::cout << test_input[j] << " ";
		std::cout << "\t";
		std::cout << "Model output: ";
		for (int j = 0; j < output.size(); j++) std::cout << output[j] << " ";
		std::cout << "\t";
		std::cout << "True data: ";
		for (int j = 0; j < output.size(); j++) std::cout << true_data[j] << " ";
		std::cout << std::endl;
	}
	std::cout << std::endl << "\033[1;31mLoss: " << finalLoss << "\033[0m" << std::endl;

	// Check if Loss meet the requirement
	double reqLoss = 0;
	if (DATASET_NAME == "cosine" || DATASET_NAME == "human_age") reqLoss = 0.01;
	else if (DATASET_NAME == "iris") reqLoss = 0.015;
	else if (DATASET_NAME == "cancer_type" || DATASET_NAME == "housing") reqLoss = 0.04;

	if (finalLoss >= reqLoss) std::cout << "\033[1;33mWarning: Loss does not meet the requiremnt(" << finalLoss << ">=" << reqLoss << ") !!\033[0m" << std::endl;
	else { std::cout << "\033[1;32mpass!!\033[0m"; }

	return 0;
}

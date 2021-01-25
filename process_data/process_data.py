import os
import numpy as np
import matplotlib.pyplot as plt


def key_func(filename):
    try:
        num = int(filename.split("_")[1])
    except IndexError as e:
        return 0

    return num


if __name__ == "__main__":

    DATA = "../data/"
    P = ["small", "medium", "large"]

    # Read in the data
    files = [file for file in os.listdir(DATA) if '.txt' in file]
    files.sort(key=key_func)
    file_data = {}
    for file in files:
        with open(DATA+file, "r") as f:
            file_data[file] = f.read()

    # string manipulate and take average of times
    for k, v in file_data.items():
        split_v = v.split("\n")
        new_v = []
        for value in split_v:
            split_val = value.split("m")

            if len(split_val) > 1:
                min = int(split_val[0]) * 60
                sec = float(split_val[1].split('s')[0]) + min
                new_v.append(sec)

        file_data[k] = np.mean(new_v)

    # Store the different P values
    small = {}
    medium = {}
    large = {}

    data = [small, medium, large]

    for k, v in file_data.items():
        if P[0] in k:
            small[k] = v
        elif P[1] in k:
            medium[k] = v
        elif P[2] in k:
            large[k] = v

    # Retrieve the values from the dict
    one = [val for val in data[0].values()]
    two = [val for val in data[1].values()]
    three = [val for val in data[2].values()]

    # Calculate the speedup
    one = (one[0] / np.array(one))[1:]
    two = (two[0] / np.array(two))[1:]
    three = (three[0] / np.array(three))[1:]

    thread_count = [1, 2, 4, 6, 8, 10, 12]

    plt.figure(figsize=[12, 8])

    plt.plot(thread_count, one, color="blue", marker="s", label="small")
    plt.plot(thread_count, two, color="red", marker="d", label="medium")
    plt.plot(thread_count, three, color="orange", marker="v", label="large")
    plt.title("Number of Threads vs. Speedup")
    plt.legend()
    plt.ylabel("Speedup")
    plt.xlabel("Number of Threads (N)")
    plt.grid(which="major", axis="y")

    plt.show()

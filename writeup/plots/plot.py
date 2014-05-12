import numpy as np
import matplotlib as mpl
font = {'weight' : 'bold',
        'size'   : 12}

mpl.rc('font', **font)
#mpl.use('Agg')
import matplotlib.pyplot as plt
import time
import random
import datetime
import csv
import sys 

csvfile = "SWEEP.csv"

def rpc_performance():
    """
    RPC performance per time
    """
    increment = 1
    times = {} # n => times
    k = 0
    alfa = 0
    with open(csvfile, 'rb') as csvf:
        csvreader = csv.reader(csvf, delimiter=',')
        for row in csvreader:
            stamp = int(row[1])
            k = int(row[2])
            alfa = int(row[3])
            n = int(row[4])
            
            if not n in times:
                times[n] = []
            times[n].append(stamp)
            
    fig = plt.figure()
    ax1 = fig.add_subplot(111)
    colors = ['blue', 'red', 'green', 'purple', 'black', 'grey']
    plots = []
    labels = []
    
    ns = times.keys()
    ns.sort()
    print "Ns = %s" % ns
    
    for j, n in enumerate(ns):
            
        starting = times[n][0]
        ending = times[n][-1] + 1
        next = starting + increment
                
        i = 1
        bins = []
        while next < ending:
            count = 0
            while times[n][i] < next:
                count += 1
                i += 1
            bins.append(count)
            starting = next
            next += increment
            
        bins = np.array(bins) / float(n)
            
        # moving average
        smoothed = [float(bins[0])]
        alpha = 0.1
        for i in range(1, len(bins), 1):
            smooth = alpha * float(bins[i-1]) + (1.0 - alpha) * float(smoothed[i-1])
            smoothed.append(smooth) 
        assert len(smoothed) == len(bins)
		
        # plot and save to disk
        color = colors[j % len(colors)]
        plot, = ax1.plot(range(len(smoothed)), smoothed, color=color)
        plots.append(plot)
        labels.append("n = %d" % n)
    
    ax1.legend(plots, labels, loc='lower right')
    ax1.set_xlabel("Time (seconds)")
    ax1.set_ylabel("RPCs per second per node")
    ax1.set_title("Performance Graph of Peerchat")
    plt.savefig("performance-sweep.png")    
    return times
    
    
times = rpc_performance()
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

csvfile = sys.argv[1]

def rpc_performance():
	"""
	RPC performnace per time
	"""
	increment = 1
	times = []
	k = 0
	alfa = 0
	with open(csvfile, 'rb') as csvf:
		csvreader = csv.reader(csvf, delimiter=',')
		for row in csvreader:
			stamp = int(row[1])
			times.append(stamp)
			k = int(row[2])
			alfa = int(row[3])
			
	starting = times[0]
	ending = times[-1] + 1
	next = starting + increment
			
	i = 1
	bins = []
	while next < ending:
		count = 0
		while times[i] < next:
			count += 1
			i += 1
		bins.append(count)
		starting = next
		next += increment
		
	# moving average
	smoothed = [float(bins[0])]
	alpha = 0.1
	for i in range(1, len(bins), 1):
		smooth = alpha * float(bins[i-1]) + (1.0 - alpha) * float(smoothed[i-1])
		smoothed.append(smooth)	
	assert len(smoothed) == len(bins)
	
	# plot and save to disk
	fig = plt.figure()
	ax1 = fig.add_subplot(111)
	plot = ax1.plot(range(len(smoothed)), smoothed)
	ax1.set_xlabel("Time (seconds)")
	ax1.set_ylabel("RPCs per second")
	ax1.set_title("Performance Graph of Peerchat")
	ax1.set_ylim((0, 20))
	plt.savefig("performance-k-%d-a-%d.png" % (k, alfa))	
	
	return bins, smoothed
	
	
bins, smoothed = rpc_performance()
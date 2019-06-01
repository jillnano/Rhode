#!usr/bin/python
#coding=utf-8

import sys
reload(sys)
sys.setdefaultencoding('utf8')
import os
import cv2
import time
from glob import glob

pyPath = os.path.dirname(os.path.realpath(__file__))

if not os.path.isdir(os.path.join(pyPath, 'target')):
	os.makedirs(os.path.join(pyPath, 'target'))
if not os.path.isdir(os.path.join(pyPath, 'stop')):
	os.makedirs(os.path.join(pyPath, 'stop'))
if not os.path.isdir(os.path.join(pyPath, 'template')):
	os.makedirs(os.path.join(pyPath, 'template'))

def findPos(template, target):
	target_rgb = cv2.imread(target)
	t_h, t_w = target_rgb.shape[0:2]
	target_gray = cv2.cvtColor(target_rgb, cv2.COLOR_BGR2GRAY)
	template_rgb = cv2.imread(template, 0)
	res = cv2.matchTemplate(target_gray, template_rgb, cv2.TM_CCOEFF_NORMED)
	value = cv2.minMaxLoc(res)
	rate = value[1]
	print rate
	if rate > 0.9:
		return (value[-1][0] + (t_w / 2), value[-1][1] + (t_h / 2))
	return None

def execute(cmd):
	print cmd
	os.system(cmd)

def click(device_id, pos):
	cmd = 'adb -s %s shell input tap %s %s'%(device_id, pos[0], pos[1])
	execute(cmd)

def screenshot(device_id, tag):
	cmd = 'adb -s %s shell screencap -p /sdcard/rhode_%s.jpg'%(device_id, tag)
	execute(cmd)
	cmd = 'adb -s %s pull /sdcard/rhode_%s.jpg %s/template/rhode_%s.jpg'%(device_id, tag, pyPath, tag)
	execute(cmd)
	return '%s/template/rhode_%s.jpg'%(pyPath, tag)

def getDeviceId():
	cmd = 'adb devices'
	out = os.popen(cmd).read()
	deviceList = out.split('\n')
	deviceList = filter(lambda x: x != '' and 'List of devices' not in x, [i.split('\tdevice')[0].strip() for i in deviceList])
	return deviceList[0]

def lock(device_id):
	cmd = 'adb -s %s shell input keyevent 26'%device_id
	execute(cmd)

def main():
	device_id = getDeviceId()
	print 'android: %s'%device_id
	target_list = glob('target/*.jpg')
	target_list.sort()
	stop_list = glob('stop/*.jpg')
	stop_list.sort()
	close = False
	while not close:
		for tar in target_list:
			template = screenshot(device_id, 'template')
			print tar
			pos = findPos(template, tar)
			while pos is None:
				template = screenshot(device_id, 'template')
				print tar
				pos = findPos(template, tar)
				time.sleep(3)
			click(device_id, pos)
			time.sleep(3)
			stop = screenshot(device_id, 'stop')
			for st in  stop_list:
				ret = findPos(stop, st)
				if ret:
					close = True
					break
			if close:
				break
	lock(device_id)

if __name__ == '__main__':
	main()
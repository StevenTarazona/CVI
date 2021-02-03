#Original code taken from https://pythonprogramming.net/opengl-rotating-cube-example-pyopengl-tutorial/

import pygame
from pygame.locals import *

from OpenGL.GL import *
from OpenGL.GLU import *
from math import * 
import math

verticies = (
    #body
    (0.25, -0.25, -1),
    (0.25, 0.25, -1),
    (-0.25, 0.25, -1),
    (-0.25, -0.25, -1),
    (0.25, -0.25, 1),
    (0.25, 0.25, 1),
    (-0.25, -0.25, 1),
    (-0.25, 0.25, 1),

    #front glass
    (0.2, 0.2, 1),
    (0.2, -0.15, 1),
    (-0.2, -0.15, 1),
    (-0.2, 0.2, 1),

    #back glass
    (0.2, 0.2, -1),
    (0.2, 0, -1),
    (-0.2, 0, -1),
    (-0.2, 0.2, -1),

    #side glass
    (-0.25, 0.2, 0.95),
    (-0.25, 0.2, -0.95),
    (-0.25, 0, -0.95),
    (-0.25, 0, 0.4),
    (-0.25, -0.15, 0.95),

    
    (0.25, 0.2, 0.95),
    (0.25, 0.2, -0.95),
    (0.25, 0, -0.95),
    (0.25, 0, 0.4),
    (0.25, -0.15, 0.95)

    )

edges = (
    #body
    (0,1),
    (0,3),
    (0,4),
    (2,1),
    (2,3),
    (2,7),
    (6,3),
    (6,4),
    (6,7),
    (5,1),
    (5,4),
    (5,7),

    #front glass
    (8,9),
    (9,10),
    (10,11),
    (11,8),

    #back glass
    (12,13),
    (13,14),
    (14,15),
    (15,12),

    #side glass
    (16,17),
    (17,18),
    (18,19),
    (19,20),
    (20,16),

    (21,22),
    (22,23),
    (23,24),
    (24,25),
    (25,21)
    )


pi = math.pi
def PointsInCircum(r,n=100):
    return [(math.cos(2*pi/n*x)*r,math.sin(2*pi/n*x)*r) for x in range(0,n)]

n=20
circle = PointsInCircum(0.1,n)
k = i = len(verticies)
for j in range(12):
    while i<k+n-1:
        edges+=((i,i+1),)
        if j % 2 == 0:
            edges+=((i,i+n),)
        i+=1
    edges+=((k,i),)
    if j % 2 == 0:
            edges+=((i,i+n),)
    k+=n
    i=k
    
verticies+=tuple(map(lambda x: (0.3,x[0]-0.25,x[1]-0.7), circle))
verticies+=tuple(map(lambda x: (0.2,x[0]-0.25,x[1]-0.7), circle))
verticies+=tuple(map(lambda x: (-0.3,x[0]-0.25,x[1]-0.7), circle))
verticies+=tuple(map(lambda x: (-0.2,x[0]-0.25,x[1]-0.7), circle))
#
verticies+=tuple(map(lambda x: (0.3,x[0]-0.25,x[1]), circle))
verticies+=tuple(map(lambda x: (0.2,x[0]-0.25,x[1]), circle))
verticies+=tuple(map(lambda x: (-0.3,x[0]-0.25,x[1]), circle))
verticies+=tuple(map(lambda x: (-0.2,x[0]-0.25,x[1]), circle))
#
verticies+=tuple(map(lambda x: (0.3,x[0]-0.25,x[1]+0.7), circle))
verticies+=tuple(map(lambda x: (0.2,x[0]-0.25,x[1]+0.7), circle))
verticies+=tuple(map(lambda x: (-0.3,x[0]-0.25,x[1]+0.7), circle))
verticies+=tuple(map(lambda x: (-0.2,x[0]-0.25,x[1]+0.7), circle))

def Cube():
    glBegin(GL_LINES)
    glColor3f(1,0,0)
    for edge in edges:
        for vertex in edge:
            glVertex3fv(verticies[vertex])
    glEnd()


def main():
    pygame.init()
    display = (800,600)
    pygame.display.set_mode(display, DOUBLEBUF|OPENGL)

    gluPerspective(45, (display[0]/display[1]), 0.1, 50.0)

    glTranslatef(0.0,0.0, -5)

    while True:
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                pygame.quit()
                quit()

        glRotatef(1, 0, 1, 0)
        glClear(GL_COLOR_BUFFER_BIT|GL_DEPTH_BUFFER_BIT)
        Cube()
        pygame.display.flip()
        pygame.time.wait(10)


main()
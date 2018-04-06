# IntelliTrivia

Intelli Trivia will process images from HQ Trivia and provide the best answer possible

## Overview

HQ screenshots have been provided in leu of a live stream from the HQ Trivia app. this could be added in by using software such as `Reflector 3` or other such mirroring services. When running the application it will parse the screenshot using Google Vision API and then use these results to query for the best answer
Answers are calculated by querying Bing for each question + answer and working out which has the most results.

## Setup

1. Create a Google Developer account
2. Generate a new Key (Service Account)
3. Setup Vision API To work with your new app
4. Copy JSON file into the base of this repo (/)
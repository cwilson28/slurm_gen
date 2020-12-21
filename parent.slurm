#!/bin/bash
#SBATCH --job-name=pipeline
#SBATCH --partition=dgx
#SBATCH --mail-type=BEGIN,END,FAIL
#SBATCH --mail-user=
#SBATCH --output=pipeline_%j.log


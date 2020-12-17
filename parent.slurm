#!/bin/bash
#SBATCH --job-name=pipeline
#SBATCH --partition=dgx
#SBATCH --mail-type=BEGIN,END,FAIL
#SBATCH --mail-user=cwilson@mdibl.org
#SBATCH --output=pipeline_%j.log

srun --input=none -N1 -c1 --tasks-per-node=1 -w dgx --mem-per-cpu=64 ./fastqc.sh
srun --input=none -N1 -c1 --tasks-per-node=1 -w dgx --mem-per-cpu=64 ./trimgalore.sh

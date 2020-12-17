#!/bin/bash
#SBATCH --job-name=fastqc
#SBATCH --partition=dgx
#SBATCH --mail-type=BEGIN,END,FAIL
#SBATCH --mail-user=cwilson@mdibl.org
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=8
#SBATCH --mem=128
#SBATCH --output=fastqc_%j.log

echo "====================================================="
pwd; hostname; date
echo "====================================================="

# Load the singularity module
module load singularity
export LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

singularity run \
--bind /home/cwilson/compbio/:/compbio \
/home/cwilson/compbio/docker/fastqc_v0_11_9.sif \
fastqc \
--outdir /compbio/analysis/JimCoffman/jcoffman_006.abOnly_cortisol_2018 \
/compbio/data/JimCoffman/jcoffman_006.abOnly_cortisol_2018/0AB1_S1_LALL_R1_001.fastq.gz \
/compbio/data/JimCoffman/jcoffman_006.abOnly_cortisol_2018/0AB1_S1_LALL_R2_001.fastq.gz \

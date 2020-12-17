#!/bin/bash
#SBATCH --job-name=trinity
#SBATCH --partition=dgx
#SBATCH --mail-type=BEGIN,END,FAIL
#SBATCH --mail-user=cwilson@mdibl.org
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=48
#SBATCH --mem=512
#SBATCH --output=trinity_%j.log

echo "====================================================="
pwd; hostname; date
echo "====================================================="

# Load the singularity module
module load singularity
export LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

singularity run \
--bind /home/cwilson/compbio/:/compbio \
/home/cwilson/compbio/docker/trinity_v_latest.sif \
trinity \
--seqType fq \
--normalize_by_read_set \
--normalize_max_read_cov 30 \
--samples_file /compbio/data/MINOTA/minota_005.coffman_rnaseq/jcoffman_005.samples.txt \
--trimmomatic \
--include_supertranscripts \
--max_memory 512G \
--CPU 48 \
--output /compbio/analysis/MINOTA/minota_005.coffman_rnaseq/trinity_run_2_supertranscript_2 \

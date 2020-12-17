#!/bin/bash

singularity run \
--bind /home/cwilson/compbio/:/compbio \
/home/cwilson/compbio/docker/trimgalore_v0_6_5.sif \
trim_galore \
--fastqc_args '"--noextract"' \
--illumina \
--paired \
--retain_unpaired \
--length_1 35 \
--length_2 35 \
--stringency 1 \
--length 20 \
--quality 20 \
--output_dir /compbio/analysis/JimCoffman/jcoffman_006.abOnly_cortisol_2018 \
/compbio/data/JimCoffman/jcoffman_006.abOnly_cortisol_2018/0AB1_S1_LALL_R1_001.fastq.gz \
/compbio/data/JimCoffman/jcoffman_006.abOnly_cortisol_2018/0AB1_S1_LALL_R2_001.fastq.gz \

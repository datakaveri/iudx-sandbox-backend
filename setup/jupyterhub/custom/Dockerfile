# Use Python 3.12 slim with explicit tag (requirement for BinderHub)
FROM python:3.12-slim

# Set up non-root user as required by BinderHub
ARG NB_USER=jovyan
ARG NB_UID=1000
ENV USER=${NB_USER}
ENV NB_USER=${NB_USER}
ENV NB_UID=${NB_UID}
ENV HOME=/home/${NB_USER}

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    libffi-dev \
    libssl-dev \
    libgeos-dev \
    libproj-dev \
    libgdal-dev \
    g++ \
    gdal-bin \
    curl \
    wget \
    git \
    --no-install-recommends && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Update C env vars so compiler can find gdal
ENV CPLUS_INCLUDE_PATH=/usr/include/gdal
ENV C_INCLUDE_PATH=/usr/include/gdal

# Create the non-root user
RUN adduser --disabled-password \
    --gecos "Default user" \
    --uid ${NB_UID} \
    ${NB_USER}

# Install Python packages including required JupyterHub components
RUN python3 -m pip install --no-cache-dir \
    notebook \
    jupyterlab \
    jupyterhub \
    numpy \
    scipy \
    matplotlib \
    pandas \
    gdal==3.6.2 \
    geopandas \
    rasterio \
    folium \
    rioxarray \
    leafmap \
    whitebox \
    seaborn \
    pysheds \
    plotly \
    dash \
    dash-leaflet \
    git+https://github.com/datakaveri/gdi-python-sdk.git

# Set environment variables for GDAL
ENV GDAL_DATA=/usr/share/gdal
ENV PROJ_LIB=/usr/share/proj

# Copy repository contents to HOME and set ownership (BinderHub requirement)
COPY . ${HOME}
USER root
RUN chown -R ${NB_UID} ${HOME}

# Switch to non-root user
USER ${NB_USER}
WORKDIR ${HOME}

# Set the command that BinderHub will run
CMD ["jupyterhub-singleuser"]
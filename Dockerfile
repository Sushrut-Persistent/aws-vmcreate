FROM registry.access.redhat.com/ubi8/ubi-minimal

# Labels
LABEL name="aws-vmcreate" \
    maintainer="sushrut.com" \
    vendor="sushrut" \
    version="1.0.0" \
    release="1" \
    summary="This service enables provisioning/de-provisioning of AWS cloud vms." \
    description="This service enables provisioning/de-provisioning AWS cloud vms."

# copy code to the build path
USER root
WORKDIR /opt
RUN chgrp -R 0 /opt && \
    chmod -R g=u /opt && \
    chmod +x -R /opt


USER 1001
ENV ec2_tag_key "Name"
ENV ec2_tag_value "Default"
ENV ec2_command "create"
#can be "delete" also
# ENV ec2_instance_type "t2.micro"
# ENV ec2_image_id "ami-0d0ca2066b861631c"

# COPY go.* ./
COPY aws-vmcreate aws-vmcreate
# RUN go mod download
# RUN go build -o aws-vmcreate
# CMD ["bash","-c","/opt/aws-vmcreate -c $ec2_command -n $ec2_tag_key -v $ec2_tag_value -ii $ec2_image_id -it $ec2_instance_type "]
CMD ["bash","-c","/opt/aws-vmcreate -c $ec2_command -n $ec2_tag_key -v $ec2_tag_value "]
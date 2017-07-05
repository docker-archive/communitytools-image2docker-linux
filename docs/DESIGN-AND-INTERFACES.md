# Application Design and Component Interfaces

This tool is a fan-out / fan-in workflow for transforming a single virtual machine image or other source material into a Docker image build context.

There are four main components to this project. 

1. CLI and Workflow Orchestrator - This is the v2c program itself. It is a native program that orchestrates plugins and the flow of data through a pipeline. Plugins are packaged as Docker images and run in containers. This program has a dependency on a running Docker engine and will use configuration from the environment like most other Docker clients.
2. Packager Plugins - At runtime a single packager plugin is selected from those installed (images pulled to the engine). The packager is run before any other plugin and makes the contents of the input material available to other plugins via volume.
3. Detective Plugins - The orchestrator starts all available detective plugins in parallel once the packager has made the input available. That input is shared with each detective at a volume mounted into each container at /v2c/disk. The role of a detective is to determine if a specific artifact or set of artifacts are present in the input material, to collect relevant sources, and to signal to the orchestrator that these artifacts should be provisioned.
4. Provisioner Plugins - If a detective signals that its target has been detected, the orchestrator will start the provisioner associated with that detective. The output from the detective is made available to the provisioner (which has no direct access to the input material). The purpose of the provisioner is to generate a Dockerfile fragment and collect relevant artifacts that should be included in the build context.

The orchestrator performs a final Dockerfile preparation phase once all provisioners have finished and the resulting material has been collected. The unpackaged input material is then discarded unless otherwise specified. Since packagers can take some time to process full disk images it can be helpful to skip this step and operate on that cache in iterative work.  

### A Note Regarding Plugins

Plugins are third-party code that might mutate without notification. As such they should always be adopted with caution and are treated by this project as potentially hostile. This is the reason that each is run in a separate container, neither packagers nor detectives have network access, detectives have read-only access to the input material, and provisioners have no direct access to the input material whatsoever.

Even with those cautions in place it is possible to cause operational harm to the machine where these programs are run. For that reason it is best to do so in a controlled environment with limited access to sensitive materials and no impact to systems in your critical path.

Please use this tool with caution.

## Packagers

Packagers accept a disk image as a volume at /input/input.vmdk, and make the extracted material available in another volume mounted at /v2c/disk. Once a packager has finished extracting the material for detection it should terminate with code 0. If a packager returns a different status code then processing will hault. 

If this program determines that material has already been extracted and cleanup was surpressed then the packaging phase will be skipped.

## Detectives

Every detective receives the contents of the VMDK at /v2c/disk as a read-only volume. A detective can signal that provisioning should occur if the contained program exits with a status code of 0. If the detective must pass material from the source image to the detective's associated provisioner then it should write that material to STDOUT. The orchestrator will buffer all data sent to STDOUT and push it to the associated provisioner's STDIN.

This stream interface can be very flexible. Some detective/provisioner pairs may not require such communication, others might only need to exchange some raw configuration such as a list of package names. Other pairs might require passing collections of files. Archives like TAR files work well in those situations.

Detectives have no network access. Detectives should not block on data from STDIN (as none will be sent).

Images that contain detectives are recognized by setting the following image labels:

* com.docker.v2c.component=detective
* com.docker.v2c.component.category=&lt;the category&gt;
* com.docker.v2c.component.description=&lt;a brief description of the detective&gt;
* com.docker.v2c.component.rel=&lt;the full repository identifier of the related provisioner&gt;

If no detectives signal a successful detection to the orchestrator then processing will hault.

## Provisioners

No provisioners will be started until all detectives have been completed. Once that has happened all targeted provisioners will be started in parallel and the STDOUT buffers of each detective will be written to the STDIN for its associated provisioner. Provisioners will not have any volumes mounted, or host port mappings made available. However, at the time of this writing, provisioners do have bridge network access.

The primary purpose of a provisioner is to prepare a Dockerfile fragment and any other related material that might be required during image assembly. To do so, a provisioner must write an uncompressed TAR stream to STDOUT. The Dockerfile fragment must be in a file named "Dockerfile" and located at the base directory of the TAR stream (I.E. "./Dockerfile").

The included Dockerfile will be analyzed and merged with the output from other Provisioners. The TAR stream sent to STDOUT will be collected and added to the resulting Dockerfile with an ADD instruction. The ADD instruction automatically extracts TAR archives based at /. So, when a provisioner constructs its archive it should base all included paths from / within the container.

With all of this composition it might seem like there is a significant opportunity for file conflicts. However, since this is a "lift and shift" project, it is anticipated that the primary source for any files included in the resulting archive originated from the source material. For that reason, even if two provisioners write to the same file they will likely be writing the same contents. That being stated, there are always edge cases and so provisioners are labeled with a category. 

Final assembly is orchestrated by processing known categories of results in a specific order. Race conditions exist between any two provisioners in the same category, but application will always overwrite os, config will overwrite application, and init will overwrite them all.

1. os - only allowed to contribute a single FROM Dockerfile instruction and no STDOUT TAR streams are processed
2. application - allowed to contribute most Dockerfile instructions and may contribute other files via TAR stream on STDOUT
3. config - allowed to contribute most Dockerfile instructions and may contribute other files via TAR stream on STDOUT
4. init - allowed to contribute ENTRYPOINT and CMD Dockerfile instructions as well as other files via TAR on STDOUT

Provisioners are identified with the following labels:

* com.docker.v2c.component=provisioner
* com.docker.v2c.component.category=&lt;category&gt;
* com.docker.v2c.component.description=&lt;a short description&gt;

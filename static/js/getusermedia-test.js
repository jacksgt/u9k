/* This JavaScript code uses ECMAScript 2017 */

/* https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Strict_mode */
'use strict';

function videoAudioTestWidget() {
    // buttons
    const start = document.getElementById("startbuttons");
    const stop = document.getElementById("stopbuttons");
    const startVideoButton = document.querySelector("#startVideoButton");
    const startAudioButton = document.querySelector("#startAudioButton");
    const svgRecordingButton = document.getElementById('svg-placeholder-image').contentDocument.getElementById("svgRecordingButton");
    const startScreenButton = document.querySelector("#startScreenButton");
    const stopButton = document.querySelector("#stopButton");
    const downloadButton = document.querySelector("#downloadButton");

    const downloadWrapper = document.querySelector("#download-wrapper");
    const message = document.getElementById("message");
    const content = document.getElementById("content");
    const overlay = document.querySelector("#video-overlay");
    const fps = overlay.querySelector("#overlay-bottom-left-text");
    const audioDuration = document.querySelector("#duration");
    const videoDuration = overlay.querySelector("#overlay-top-right-text");
    const resolution = overlay.querySelector("#overlay-bottom-right-text");
    const placeholder = document.querySelector("#svg-placeholder-image");

    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
        message.innerHTML = "Sorry, your browser does not support MediaDevices Web API.<br>Are you accessing this site via an unencrypted connection?";
        start.style.display = "none";
        return;
    }

    // multimedia elements for playing recorded content
    const video = document.createElement("video");
    video.setAttribute("width", "100%");

    const audio = document.createElement("audio");
    audio.setAttribute("controls", true);

    // variables for FPS calculations (Firefox only)
    let fps_now = 0;
    let fps_total = 0;
    let last_fps_time;
    let first_fps_time;
    let last_fps_frames;
    let first_fps_frames;

    /* define and register button event handlers */
    function startScreen() {
        const stream = startMedia("getDisplayMedia", {video: true});
        startRecording(stream);
        recordingType = 'video/webm';
    }
    startScreenButton.addEventListener('click', startScreen);

    function startAudio() {
        const stream = startMedia("getUserMedia", {audio: true});
        startRecording(stream);
        recordingType = 'audio/ogg; codecs=opus';
    }
    startAudioButton.addEventListener('click', startAudio);

    function toggleRecording() {
        if (recordingDuration == -1) {
            startAudio();
        } else {
            stopMedia();
        }
    }
    svgRecordingButton.addEventListener('click', toggleRecording);

    function startVideo() {
        const stream = startMedia("getUserMedia", {video: true});
        startRecording(stream);
        recordingType = 'video/webm';
    }
    startVideoButton.addEventListener('click', startVideo);

    let recordedChunks = [];
    let mediaRecorder = null;
    let recordingType = '';
    let recordingDuration = -1;
    async function startRecording(streamPromise) {
        // clear array of recorded chunks
        recordingDuration = 0.0;
        recordedChunks = [];

        const stream = await streamPromise;
        if (! stream) {
            console.error("Cannot start recording for stream:", stream);
            return;
        }

        mediaRecorder = new MediaRecorder(stream);

        mediaRecorder.addEventListener('dataavailable', function(e) {
            recordedChunks.push(e.data);
        });
        mediaRecorder.start();
    }

    function stopRecording() {
        recordingDuration = -1;
        return new Promise(resolve => {
            mediaRecorder.addEventListener("stop", () => {
                const dataBlob = new Blob(recordedChunks, { 'type' : recordingType });
                const dataUrl = window.URL.createObjectURL(dataBlob);
                resolve({dataUrl, dataBlob});
            });

            mediaRecorder.stop();
        });
    }

    async function stopMedia() {
        if (video.srcObject) {
            for (const track of video.srcObject.getTracks()) {
                track.stop();
            }
            // clear old content of video element
            video.srcObject = null;

            const {dataUrl, dataBlob} = await stopRecording();
            console.log("VIDEO URL", dataUrl);

            // disable overlay
            overlay.style.display = "none";

            // add new content for video element
            video.src = dataUrl;
            video.setAttribute("controls", true);
            video.play();

            // enable download button
            downloadButton.onclick = function() {
                downloadBlobAsFile(dataUrl, "recording.webm");
                return false;
            };
        } else if (audio.srcObject) {
            for (const track of audio.srcObject.getTracks()) {
                track.stop();
            }
            const {dataUrl, dataBlob} = await stopRecording();
            console.info("AUDIO URL", dataUrl);

            // clear old content of audio element
            audio.srcObject = null;

            // add new content
            content.innerHTML = "";
            content.appendChild(audio);
            audio.src = dataUrl;
            audio.muted = false; // make sure sound is on
            audio.play();

            // enable download button
            downloadButton.onclick = function() {
                downloadBlobAsFile(dataUrl, "recording.ogg");
                return false;
            };
        }
        stop.style.display = "none";
        // start.style.display = "block";
        downloadWrapper.style.display = "block";
    }
    stopButton.addEventListener('click', stopMedia);

    // function pauseMedia() {
    //     if (saved_stream) {
    //         if (saved_stream.getVideoTracks().length) {
    //             video.srcObject = saved_stream;
    //             video.play();
    //         } else {
    //             audio.srcObject = saved_stream;
    //             audio.play();
    //         }
    //         saved_stream = null;
    //     } else {
    //         if (video.srcObject) {
    //             video.pause();
    //             saved_stream = video.srcObject;
    //             video.srcObject = null;
    //         } else if (audio.srcObject) {
    //             audio.pause();
    //             saved_stream = audio.srcObject;
    //             audio.srcObject = null;
    //         }
    //     }
    // }

    const wait = ms => new Promise(resolve => setTimeout(resolve, ms));

    async function startMedia(gum, constraints) {
        stop.style.display = "block";
        start.style.display = "none";
        try {
            message.innerHTML = "Please grant permission for camera / microphone.";
            const stream = await navigator.mediaDevices[gum](constraints);
            message.innerHTML = "";
            if (stream.getVideoTracks().length) {

                // enable video overlay, disable placeholder and audio duration, add video object
                overlay.style.display = "block";
                placeholder.style.display = "none";
                audioDuration.style.display = "none";
                content.appendChild(video);

                // start playing the video
                video.srcObject = stream;
                video.play();
                setTimeout(showDuration, 1000);

                // https://developer.mozilla.org/en-US/docs/Web/API/MediaStreamTrack
                // https://developer.mozilla.org/en-US/docs/Web/API/MediaTrackSettings

                video.addEventListener('loadedmetadata', function(e){
                    showResolution(video.videoWidth, video.videoHeight);
                    // https://developer.mozilla.org/en-US/docs/Web/API/CSS_Object_Model/Determining_the_dimensions_of_elements
                    overlay.style.width = `${video.offsetWidth}px`;
                    overlay.style.height = `${video.offsetHeight}px`;
                });

                first_fps_time = last_fps_time = new Date();
                first_fps_frames = last_fps_frames = video.mozPaintedFrames;
                setTimeout(get_fps, 1000);
            } else {
                audio.muted = true; // prevent nasty echo
                audio.srcObject = stream;
                audio.play();
                setTimeout(showDuration, 1000);
            }
            return stream;
        } catch (err) {
            console.error(err);
            message.innerHTML = `<p class='error'>${err}</p>`;
            stopMedia();
            return null;
        }
    }

    // adapted from https://randomtutes.com/2019/08/02/download-blob-as-file-in-javascript/
    function downloadBlobAsFile(blob_url, filename){
        const contentType = 'application/octet-stream';

        const a = document.createElement('a');
        a.href = blob_url;
        a.download = filename;
        a.dataset.downloadurl =  [contentType, a.download, a.href].join(':');
        a.click();
    }

    function showDuration(element) {
        if (recordingDuration == -1) return;

        // refresh duration every second
        setTimeout(showDuration, 1000);

        // formats a number with zero comma digits and leading 0
        function num_fmt(num) {
            let s = num.toFixed(0);
            if (num < 10) {
                s = "0" + s;
            }
            return s;
        }

        const minutes = num_fmt(Math.floor(recordingDuration / 60));
        const seconds = num_fmt(recordingDuration % 60);
        const duration_str = `${minutes}:${seconds}`;
        audioDuration.innerHTML = duration_str;
        videoDuration.innerHTML = duration_str;
        recordingDuration += 1;
    }

    function showResolution(width, height) {
        if (! width || ! height) {
            console.warn("Got invalid resolution, showing nothing:", width, height);
            resolution.innerHTML = "";
            return;
        }

        resolution.innerHTML = `${width}x${height}`;
    }

    function fps_format(fps_now, fps_total) {
        if (Number.isNaN(fps_now)) {
            return "";
        }

        return `${fps_now.toFixed(0)} FPS`;
    }

    function get_fps() {
        if (recordingDuration == -1) return;

        // refresh FPS value every second
        setTimeout(get_fps, 1000);

        // Note: mozPaintedFrames only works on Gecko-based browsers (Firefox)
        const now = new Date();
        const frames = video.mozPaintedFrames;
        fps_now = (frames - last_fps_frames)/((now - last_fps_time)/1000);
        fps_total = (frames - first_fps_frames)/((now - first_fps_time)/1000);
        fps.innerHTML = fps_format(fps_now, fps_total);
        last_fps_time = now;
        last_fps_frames = frames;
    }
}

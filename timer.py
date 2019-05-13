import json
import time
from datetime import datetime
import os
import strictyaml as yaml
from pynput import keyboard
import sys

# Read the config file
CONFIG_FILENAME = "config.yml"
CONFIG_SCHEMA = yaml.Map({
    "exam_name": yaml.Str(),
    "exam_section": yaml.Str(),
    "total_questions": yaml.Int(),
    "total_time": yaml.Int()
})
LOG_PREFIX = "logs"
REPORT_PREFIX = "reports"
RECORD_PREFIX = "records"

if not os.path.isdir(LOG_PREFIX):
    os.mkdir(LOG_PREFIX)

if not os.path.isdir(REPORT_PREFIX):
    os.mkdir(REPORT_PREFIX)

if not os.path.isdir(RECORD_PREFIX):
    os.mkdir(RECORD_PREFIX)


def get_config_file():
    with open(CONFIG_FILENAME, "r", encoding="utf8") as file:
        return file.read()


def sanitize(original: str, no_lower: bool = False) -> str:
    if not no_lower:
        original = original.lower()
    return original.replace(" ", "-").replace(":", "-")


config = yaml.load(get_config_file(), CONFIG_SCHEMA).data


# Timer
class Timer:
    def __init__(self, question_index: int):
        self.question_index = question_index
        self.time_ranges = []
        self._last_start = None

    def start(self):
        if not self._last_start:
            self._last_start = time.time()
        else:
            raise ValueError("Timer has already started")

    def stop(self):
        if self._last_start:
            self.time_ranges.append([self._last_start, time.time()])
            self._last_start = None
        else:
            raise ValueError("Timer has already stopped")

    def get_status(self):
        if self._last_start:
            return "RUNNING"
        else:
            return "STOPPED"

    def get_total_time(self) -> float:
        if self.get_status() == "STOPPED":
            counter = 0.0
            for start, end in self.time_ranges:
                counter += end - start
            return counter
        else:
            raise ValueError("Timer is still running")

    def get_friendly_time(self) -> str:
        t = int(self.get_total_time())
        return f"{int(t / 60)}m{t % 60}s"

    def get_detail(self) -> str:
        messages = []
        for index, (start, end) in enumerate(self.time_ranges):
            start = datetime.fromtimestamp(start)
            end = datetime.fromtimestamp(end)
            elapsed = str(end - start)
            messages.append(f"[#{index + 1}] {start.strftime('%H:%M:%S.%f')} ~ {end.strftime('%H:%M:%S.%f')} ({elapsed})")
        return messages


# Session
class Session:
    session_started = datetime.now()
    name = f"{config.get('exam_name')} - Section {config.get('exam_section')}"
    timing = f"Number of Questions: {config.get('total_questions')} | Time: {int(config.get('total_time') / 60)} hour {config.get('total_time') % 60} minutes"
    file_name = f"{sanitize(datetime.now().isoformat(), True)}_{sanitize(config.get('exam_name'))}_section-{sanitize(config.get('exam_section')).upper()}.log"
    log_location = os.path.join(LOG_PREFIX, file_name)
    report_location = os.path.join(REPORT_PREFIX, file_name.replace(".log", ".txt"))
    record_location = os.path.join(RECORD_PREFIX, file_name.replace(".log", ".json"))
    started = False
    current = 0
    timers = [Timer(index) for index in range(config.get("total_questions"))]
    jump_mode = False
    jump_candidate = []


start_message = f"""[ == Exam Timer == ]

- Current Exam Configuration
\t- Session: {Session.name}
\t- Description: {Session.timing}
\t- Logging at: {Session.log_location}

[ == Press "s" to Start Session == ]
"""


def log(content):
    if "\n" in content:
        for stripped in content.split("\n"):
            log(stripped)
        return False
    content = f"[{datetime.now().isoformat()}] {content}"
    print(content)
    with open(Session.log_location, "a+") as log_file:
        log_file.write(content + "\n")


def prog_exit():
    RETURN = '\n'
    TAB = '\t'
    DOUBLE_TAB = TAB * 2
    now = datetime.now()
    message = f"""[ == Session Ended == ]

- Summary
\t- Started at: {Session.session_started.strftime('%H:%M:%S.%f')}
\t- Ended at: {now.strftime('%H:%M:%S.%f')}
\t- Elapsed: {str(now - Session.session_started)}

- Details
{RETURN.join([f"{TAB}- Question {timer.question_index + 1}{RETURN}{RETURN.join([f'{DOUBLE_TAB}- {message}' for message in timer.get_detail()])}" for timer in Session.timers])}
"""
    dataset = []
    for timer in Session.timers:
        current_set = {
            "question": timer.question_index + 1,
            "ranges": timer.time_ranges
        }
        dataset.append(current_set)

    # write analyzable file (record)
    try:
        encoded = json.dumps(dataset)
        with open(Session.record_location, "w") as json_file:
            json_file.write(encoded)
    except:
        pass

    # write report
    with open(Session.report_location, "w") as record_file:
        record_file.write(message)

    # log them out
    log(message)

    # see you next time!
    sys.exit(0)


def on_press(key):
    try:
        k = key.char  # single-char keys
    except:
        k = key.name  # other keys
    if key == keyboard.Key.esc:
        return False  # stop listener

    # keys interested
    if k in ['s', 'space', 'j', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', 'enter', 'backspace', 'e', 'p']:
        # self.keys.append(k) # store it in global-like variable

        # Start the exam
        if k == "s":
            if not Session.started:
                log("Starting Exam; Start at Question 1")
                Session.started = True
                Session.timers[0].start()
                Session.current = 0
            else:
                raise AttributeError("Session has already started")
        elif k == "space":  # Next question
            if Session.started:
                Session.timers[Session.current].stop()  # Stop previous question's timer
                log(
                    f"Time used for Question {Session.current + 1}: {Session.timers[Session.current].get_friendly_time()}")
                if config.get("total_questions") - 1 <= Session.current:
                    prog_exit()
                Session.current += 1
                Session.timers[Session.current].start()
                log(f"Starting Question {Session.current + 1}.")
            else:
                raise AttributeError("Session is not started")
        elif k == "p":  # pause/resume timer
            if Session.started:
                if Session.timers[Session.current].get_status() == "STOPPED":
                    Session.timers[Session.current].start()  # Start current question's timer
                    log(f"Resuming Question {Session.current + 1}.")
                else:
                    Session.timers[Session.current].stop()  # Stop current question's timer
                    log(
                        f"Stopped Question {Session.current + 1}. Currently elapsed: {Session.timers[Session.current].get_friendly_time()}")
            else:
                raise AttributeError("Session is not started")
        elif k == "e":  # exit program
            if Session.started:
                Session.timers[Session.current].stop()  # Stop current question's timer
                log(
                    f"Stopped Question {Session.current + 1}. Currently elapsed: {Session.timers[Session.current].get_friendly_time()}")
                prog_exit()
            else:
                raise AttributeError("Session is not started")
        elif k == "j":  # Jump to question
            if Session.started:
                if Session.jump_mode:
                    raise AttributeError("Jump mode has already been started")
                Session.timers[Session.current].stop()  # Stop previous question's timer
                log(
                    f"Time used for Question {Session.current + 1}: {Session.timers[Session.current].get_friendly_time()}")
                log("Input the question you want to jump to:")
                Session.jump_mode = True
            else:
                raise AttributeError("Session is not started")
        elif k in ['1', '2', '3', '4', '5', '6', '7', '8', '9', '0']:
            if Session.started:
                if not Session.jump_mode:
                    return False
                Session.jump_candidate.append(k)
                print(k)
            else:
                raise AttributeError("Session is not started")
        elif k == "backspace":
            if Session.started:
                if not Session.jump_mode:
                    return False
                del Session.jump_candidate[-1]
            else:
                raise AttributeError("Session is not started")
        elif k == "enter":
            if Session.started:
                if not Session.jump_mode:
                    return False
                jump_to = int("".join(Session.jump_candidate)) - 1

                Session.current = jump_to
                Session.timers[Session.current].start()
                log(f"Starting Question {Session.current + 1}.")

                Session.jump_mode = False
                Session.jump_candidate = []
            else:
                raise AttributeError("Session is not started")


if __name__ == '__main__':
    log(start_message)
    lis = keyboard.Listener(on_press=on_press)
    lis.start()  # start to listen on a separate thread
    lis.join()  # no this if main thread is polling self.keys

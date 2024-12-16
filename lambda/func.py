

import logging
import time

logger = logging.getLogger()
logger.setLevel(logging.INFO)
def lambda_handler(event, context):
    try:
        
        time.sleep(1)
        print("Hellooooo")
        return {
            "status": "SUCCESS",
            "details": "I am amr successful.",
            "processed" : event
        }
    except Exception as e:
        logger.error(e, exc_info=True)
        return {
            "status": "FAILED",
            "details": "Error while executing lambda function.",
            "error": str(e)
        }


